package messages

// Message from server to client.
const (
	AskSupportMessageType         = "askSupport"
	OrdersReceivedMessageType     = "ordersReceived"
	OrdersConfirmationMessageType = "ordersConfirmation"
)

// Message sent from server when asking a supporting player who to support in an embattled area.
type AskSupportMessage struct {
	BaseMessage
	From     string   `json:"from"`
	To       string   `json:"to"`
	Battlers []string `json:"battlers"` // List of possible players to support in the battle.
}

// Message sent from server to all clients when valid orders are received from all players.
type OrdersReceivedMessage struct {
	BaseMessage
	Orders map[string][]OrderMessage `json:"orders"` // Maps a player's ID to their submitted orders.
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmationMessage struct {
	BaseMessage
	Player string `json:"player"` // The player who submitted orders.
}
