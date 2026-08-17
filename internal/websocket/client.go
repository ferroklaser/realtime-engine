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
		err = json.Unmarshal(jsonData, &command)

		if err != nil {
			log.Printf("Error unmarshaling JSON: %s", err)
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

		case types.CommandUnsubscribe:
			unsubscribeRequest := SubscriptionRequest{
				Client:  client,
				Channel: command.Channel,
			}

			client.Hub.Unsubscribe <- unsubscribeRequest

		default:
		}
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

		err = client.Conn.WriteMessage(websocket.TextMessage, jsonData)

		if err != nil {
			return
		}
	}
}
