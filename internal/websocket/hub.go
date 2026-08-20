package websocket

import (
	"log"
	"realtime/internal/types"
)

type Hub struct {
	Clients        map[*Client]bool
	Channels       map[string]map[*Client]bool
	ClientChannels map[*Client]map[string]bool

	Broadcasts  chan types.Message
	Register    chan *Client
	UnRegister  chan *Client
	Subscribe   chan SubscriptionRequest
	Unsubscribe chan SubscriptionRequest
}

func NewHub() *Hub {
	return &Hub{
		Clients:        make(map[*Client]bool),
		Channels:       make(map[string]map[*Client]bool),
		ClientChannels: make(map[*Client]map[string]bool),

		Broadcasts:  make(chan types.Message),
		Register:    make(chan *Client),
		UnRegister:  make(chan *Client),
		Subscribe:   make(chan SubscriptionRequest),
		Unsubscribe: make(chan SubscriptionRequest),
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case request := <-hub.Subscribe:
			if _, ok := hub.Channels[request.Channel]; !ok {
				hub.Channels[request.Channel] = make(map[*Client]bool)
			}

			hub.Channels[request.Channel][request.Client] = true

			if _, ok := hub.ClientChannels[request.Client]; !ok {
				hub.ClientChannels[request.Client] = make(map[string]bool)
			}

			hub.ClientChannels[request.Client][request.Channel] = true

		case request := <-hub.Unsubscribe:

			if clients, ok := hub.Channels[request.Channel]; ok {
				delete(clients, request.Client)

				if len(clients) == 0 {
					delete(hub.Channels, request.Channel)
				}
			}

			if channels, ok := hub.ClientChannels[request.Client]; ok {
				delete(channels, request.Channel)

				if len(channels) == 0 {
					delete(hub.ClientChannels, request.Client)
				}
			}

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
