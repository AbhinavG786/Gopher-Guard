package engine

import (
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

type SlidingWindow struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	RaftNode *raft.Raft
}

func NewSlidingWindow() *SlidingWindow {
	return &SlidingWindow{
		requests: make(map[string][]time.Time),
	}
}

func (sw *SlidingWindow) Allow(key string, limit int32, windowMs time.Duration) (bool, int32, error) {
	if sw.RaftNode != nil && sw.RaftNode.State() != raft.Leader {
		return false, 0, fmt.Errorf("node is not the raft leader")
	}

	now := time.Now()
	boundary := now.Add(-windowMs)

	sw.mu.Lock()
	timestamps := sw.requests[key]
	validTimestamps := timestamps[:0]
	for _, ts := range timestamps {
		if ts.After(boundary) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	sw.requests[key] = validTimestamps
	validCount := len(validTimestamps)
	sw.mu.Unlock()

	if validCount > math.MaxInt32 {
		return false, 0, fmt.Errorf("request count overflow")
	}
	validCount32 := int32(validCount)

	// if (int32(len(validTimestamps)) < int32(limit)){
	// 	validTimestamps=append(validTimestamps, now)
	// 	sw.requests[key]=validTimestamps
	// 	remaining:=limit - int32(len(validTimestamps))
	// 	return true,remaining
	// }

	if validCount32 >= limit {
		RateLimitRequests.WithLabelValues("denied").Inc()
		return false, 0, nil
	}

	cmd := LogCommand{
		Type:      CommandAddTimestamp,
		Key:       key,
		Timestamp: now,
	}

	data, err := cmd.Encode()
	if err != nil {
		return false, 0, fmt.Errorf("failed to encode command: %w", err)
	}

	if sw.RaftNode != nil {
		future := sw.RaftNode.Apply(data, 500*time.Millisecond)
		if err := future.Error(); err != nil {
			return false, 0, err
		}
	} else {
		sw.mu.Lock()
		sw.requests[key] = append(sw.requests[key], now)
		sw.mu.Unlock()
	}

	RateLimitRequests.WithLabelValues("allowed").Inc()

	// sw.requests[key]=validTimestamps
	remaining := limit - validCount32 - 1
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining, nil
}

func (sw *SlidingWindow) StartJanitor(interval time.Duration, maxWindow time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if sw.RaftNode != nil && sw.RaftNode.State() == raft.Leader {
				slog.Info("Janitor: Node is Leader, starting cleanup")
				sw.cleanUp(maxWindow)
			} else {
				slog.Info("Janitor: Node is Follower, skipping cleanup")
			}
		}
	}()
}

func (sw *SlidingWindow) cleanUp(maxWindow time.Duration) {
	sw.mu.RLock()
	var keysToDelete []string
	now := time.Now()
	boundary := now.Add(-maxWindow)

	for key, timestamps := range sw.requests {
		// var validTimestamps []time.Time
		// for _, ts := range timestamps {
		if len(timestamps) > 0 && timestamps[len(timestamps)-1].Before(boundary) {
			keysToDelete = append(keysToDelete, key)
		}
		// if ts.After(boundary) {
		// 	validTimestamps = append(validTimestamps, ts)
		// }
	}
	sw.mu.RUnlock()
	// if len(validTimestamps) > 0 {
	// 	sw.requests[key] = validTimestamps
	// } else {
	// 	delete(sw.requests, key)
	// }
	for _, key := range keysToDelete {
		cmd := LogCommand{Type: CommandDeleteKey, Key: key}
		data, _ := cmd.Encode()
		sw.RaftNode.Apply(data, 500*time.Millisecond)
	}
}
