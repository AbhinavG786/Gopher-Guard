package engine

import (
	"sync"
	"testing"
	"time"
)

func TestSlidingWindow_Race(t *testing.T) {
	sw := NewSlidingWindow()
	key := "stress_test_user"
	limit := int32(1000)
	window := 1 * time.Second

	var wg sync.WaitGroup
	// We'll fire 500 goroutines at once
	concurrentRequests := 500

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = sw.Allow(key, limit, window)
		}()
	}
	wg.Wait()
}
