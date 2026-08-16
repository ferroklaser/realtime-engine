package types

import "encoding/json"

type CommandType string

const (
	CommandSubscribe   CommandType = "subcribe"
	CommandUnsubscribe CommandType = "unsubscribe"
	CommandMessage     CommandType = "message"
)

type Command struct {
	Type    string          `json:"type"`
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data,omitempty"`
}
