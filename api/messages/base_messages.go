package messages

// Embedded struct in all message types.
type BaseMessage struct {
	Type string `json:"type"` // Allows for correctly identifying incoming messages.
}

// Basic order message type used as part of other messages.
type OrderMessage struct {
	OrderType string `json:"orderType"`
	From      string `json:"from"`
	To        string `json:"to"`
	Via       string `json:"via"`
	Build     string `json:"build"`
}
