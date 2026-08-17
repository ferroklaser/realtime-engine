package websocket

import (
	"encoding/json"
	"log"
	"realtime/internal/types"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan types.Message
}

func (client *Client) ReadPump() {
	defer func() {
		client.Hub.UnRegister <- client
		client.Conn.Close()
	}()

	for {
		// server reads message from user
		_, jsonData, err := client.Conn.ReadMessage()

		if err != nil {
			break
		}

		var command types.Command
		err2 := json.Unmarshal(jsonData, &command)

		if err2 != nil {
			log.Printf("Error unmarshaling JSON: %s", err2)
		}

		switch command.Type {
		case types.CommandMessage:
			message := types.Message{
				Channel: command.Channel,
				Data:    command.Data,
			}

			client.Hub.Broadcasts <- message

		case types.CommandSubscribe:
			subscribeRequest := SubscriptionRequest{
				Client:  client,
				Channel: command.Channel,
			}

			client.Hub.Subscribe <- subscribeRequest
		default:
		}

		// var message types.Message
		// err2 := json.Unmarshal(jsonData, &message)

		// log.Printf("Server received: message")
		// client.Hub.Broadcasts <- message
	}
}

func (client *Client) WritePump() {
	defer func() {
		client.Conn.Close()
	}()

	for {
		message, ok := <-client.Send

		if !ok {
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		jsonData, err := json.Marshal(message)

		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
		}

		err2 := client.Conn.WriteMessage(websocket.TextMessage, jsonData)

		if err2 != nil {
			return
		}
	}
}
