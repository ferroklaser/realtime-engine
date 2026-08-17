package websocket

type SubscriptionRequest struct {
	Client  *Client
	Channel string
}
