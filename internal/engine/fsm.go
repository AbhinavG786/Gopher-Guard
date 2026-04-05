package engine

import (
	"encoding/json"
	"io"
	"log/slog"
	"time"

	"github.com/hashicorp/raft"
)

type LimiterFSM struct {
	Engine *SlidingWindow
}

type Snapshot struct {
	state map[string][]time.Time
}

func (s *Snapshot) Release() {}

func (f *LimiterFSM) Apply(logEntry *raft.Log) interface{} {
	cmd, err := DecodeCommand(logEntry.Data)
	if err != nil {
		slog.Error("Failed to decode command", slog.String("error", err.Error()))
		return nil
	}
	switch cmd.Type {
	case CommandAddTimestamp:
		f.Engine.mu.Lock()
		defer f.Engine.mu.Unlock()
		f.Engine.requests[cmd.Key] = append(f.Engine.requests[cmd.Key], cmd.Timestamp)
		slog.Debug("Applied ADD_TIMESTAMP via Raft", slog.String("key", cmd.Key))
	case CommandDeleteKey:
		f.Engine.mu.Lock()
		defer f.Engine.mu.Unlock()
		delete(f.Engine.requests, cmd.Key)
		slog.Debug("Applied DELETE_KEY via Raft", slog.String("key", cmd.Key))
	default:
		slog.Warn("Unknown command type", slog.String("type", string(cmd.Type)))
	}
	return nil
}

func (f *LimiterFSM) Snapshot() (raft.FSMSnapshot, error) {
	slog.Info("Raft requested a snapshot of the FSM")
	f.Engine.mu.Lock()
	defer f.Engine.mu.Unlock()
	clone := make(map[string][]time.Time)
	for k, v := range f.Engine.requests {
		clone[k] = append([]time.Time(nil), v...)
	}
	return &Snapshot{state: clone}, nil
}

func (f *LimiterFSM) Restore(rc io.ReadCloser) error {
	slog.Info("Restoring FSM state from snapshot")
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	newState := make(map[string][]time.Time)
	if err := json.Unmarshal(data, &newState); err != nil {
		return err
	}
	f.Engine.mu.Lock()
	defer f.Engine.mu.Unlock()
	f.Engine.requests = newState

	return nil
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	slog.Info("Persisting snapshot")
	data, err := json.Marshal(s.state)
	if err != nil {
		err := sink.Cancel()
		return err
	}

	if _, err := sink.Write(data); err != nil {
		err := sink.Cancel()
		return err
	}
	return sink.Close()
}
