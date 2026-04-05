package engine

import (
	"encoding/json"
	"time"
)

type CommandType string

const (
	CommandAddTimestamp CommandType = "ADD_TIMESTAMP"
)

type LogCommand struct {
	Type      CommandType `json:"type"`
	Key       string      `json:"key"`
	Timestamp time.Time   `json:"timestamp"`
}

func (lg *LogCommand) Encode() ([]byte, error) {
	return json.Marshal(lg)
}

func DecodeCommand(data []byte) (*LogCommand, error) {
	var cmd LogCommand
	err := json.Unmarshal(data, &cmd)
	return &cmd, err
}
