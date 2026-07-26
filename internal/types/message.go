package types

import "encoding/json"

type Message struct {
	ID      string
	Channel string
	Data    json.RawMessage
}
