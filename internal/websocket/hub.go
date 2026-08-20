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
		case client := <-hub.Register:
			hub.registerClient(client)

		case client := <-hub.UnRegister:
			hub.unregisterClient(client)

		case request := <-hub.Subscribe:
			hub.handleSubscribe(request)

		case request := <-hub.Unsubscribe:
			hub.handleUnsubscribe(request)

		case msg := <-hub.Broadcasts:
			hub.handleBroadcast(msg)
		}
	}
}

func (hub *Hub) registerClient(client *Client) {
	hub.Clients[client] = true
	log.Println("Client registered!")
}

func (hub *Hub) unregisterClient(client *Client) {
	if _, ok := hub.Clients[client]; ok {
		for channel := range hub.ClientChannels[client] {
			delete(hub.Channels[channel], client)

			if len(hub.Channels[channel]) == 0 {
				delete(hub.Channels, channel)
			}
		}
		delete(hub.ClientChannels, client)
		delete(hub.Clients, client)
		close(client.Send)
	}
}

func (hub *Hub) handleSubscribe(request SubscriptionRequest) {
	if _, ok := hub.Clients[request.Client]; !ok {
		return
	}

	if _, ok := hub.Channels[request.Channel]; !ok {
		hub.Channels[request.Channel] = make(map[*Client]bool)
	}

	hub.Channels[request.Channel][request.Client] = true

	if _, ok := hub.ClientChannels[request.Client]; !ok {
		hub.ClientChannels[request.Client] = make(map[string]bool)
	}

	hub.ClientChannels[request.Client][request.Channel] = true
}

func (hub *Hub) handleUnsubscribe(request SubscriptionRequest) {
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
}

func (hub *Hub) handleBroadcast(msg types.Message) {
	log.Printf("Hub is broadcasting")

	for client := range hub.Channels[msg.Channel] {
		select {
		case client.Send <- msg:
			//Message successful
		default:
			hub.unregisterClient(client)
		}
	}
}
