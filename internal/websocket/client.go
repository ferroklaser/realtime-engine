package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

func (client *Client) ReadPump() {
	defer func() {
		client.Hub.UnRegister <- client
		client.Conn.Close()
	}()

	for {
		// server reads message from user
		_, message, err := client.Conn.ReadMessage()

		if err != nil {
			break
		}
		log.Printf("Server received: %s", message)
		client.Hub.Broadcasts <- message
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

		err := client.Conn.WriteMessage(websocket.TextMessage, message)

		if err != nil {
			return
		}
	}
}
