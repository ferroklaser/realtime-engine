package websocket

type SubscribeRequest struct {
	Client  *Client
	Channel string
}
