package types

import "encoding/json"

type Message struct {
	ID      string          `json:"id,omitempty"` //name in the json file
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}
