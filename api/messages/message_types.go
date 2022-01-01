package messages

// Message from client to server.
const (
	SubmitOrdersMessageType = "submitOrders"
	GiveSupportMessageType  = "giveSupport"
	QuitMessageType         = "quit"
	KickMessageType         = "kick"
)

// Message from server to client.
const (
	AskSupportMessageType         = "askSupport"
	OrdersReceivedMessageType     = "ordersReceived"
	OrdersConfirmationMessageType = "ordersConfirmation"
)

// Message used for the throne expansion.
const (
	WinterVoteMessageType = "winterVote"
	SwordMessageType      = "sword"
	RavenMessageType      = "raven"
)

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

// Message sent from client when submitting orders.
type SubmitOrdersMessage struct {
	BaseMessage
	Orders []OrderMessage `json:"orders"`
}

// Message sent from client when declaring who to support with their support order.
type GiveSupportMessage struct {
	BaseMessage
	From   string `json:"from"`   // Name of the area in which the support order is placed.
	To     string `json:"to"`     // Name of the area to which the support order is going.
	Player string `json:"player"` // ID of the player in the destination area to support.
}

// Message sent from client when they want to quit the game.
type QuitMessage struct {
	BaseMessage
}

// Message sent from client when they vote to kick another player.
type KickMessage struct {
	BaseMessage
	Player string `json:"player"` // ID of the player to votekick.
}

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
	Orders map[string][]OrderMessage `json:"orders"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmationMessage struct {
	BaseMessage
	Player string `json:"player"` // The player who submitted orders.
}

// Message passed from the client during winter council voting.
// Used for the throne expansion.
type WinterVoteMessage struct {
	BaseMessage
	Player string `json:"player"` // ID of the player that the submitting player votes for.
}

// Message passed from the client with the sword to declare where they want to use it.
// Used for the throne expansion.
type SwordMessage struct {
	BaseMessage
	Area        string // Name of the area in which the player wants to use the sword in battle.
	BattleIndex int    // Index of the battle in which to use the sword, in case of several battles in the area.
}

// Message passed from the client with the raven when they want to spy on another player's orders.
// Used for the throne expansion.
type RavenMessage struct {
	BaseMessage
	Player string // ID of the player on whom to spy.
}
