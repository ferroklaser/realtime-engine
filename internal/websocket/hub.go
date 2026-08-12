package websocket

import (
	"log"
	"realtime/internal/types"
)

type Hub struct {
	Clients    map[*Client]bool
	Broadcasts chan types.Message
	Register   chan *Client
	UnRegister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcasts: make(chan types.Message),
		Register:   make(chan *Client),
		UnRegister: make(chan *Client),
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.Register:
			hub.Clients[client] = true
			log.Println("Client registered!")
		case client := <-hub.UnRegister:
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				close(client.Send)
			}
		case msg := <-hub.Broadcasts:
			log.Printf("Hub is broadcasting")
			for client := range hub.Clients {
				select {
				case client.Send <- msg:
					//Message successful
				default:
					close(client.Send)
					delete(hub.Clients, client)
				}
			}
		}
	}
}
