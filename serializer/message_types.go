package serializer

const (
	OrdersMessageType  = "orders"
	SupportMessageType = "support"
	QuitMessageType    = "quit"
	KickMessageType    = "kick"

	// Message types used for the throne expansion.
	WinterVoteMessageType = "winterVote"
	SwordMessageType      = "sword"
	RavenMessageType      = "raven"
)

type IncomingMessages struct {
	Orders     chan OrdersMessage
	Support    chan SupportMessage
	Quit       chan QuitMessage
	Kick       chan KickMessage
	WinterVote chan WinterVoteMessage
}

type BaseMessage struct {
	Type string `json:"type"`
}

func (message BaseMessage) GetType() string {
	return message.Type
}

type Message interface {
	GetType() string
}

// Message passed from the client when submitting orders.
type OrdersMessage struct {
	BaseMessage
	Orders []OrderMessage `json:"orders"`
}

type OrderMessage struct {
	Type  string `json:"type"`
	From  string `json:"from"`
	To    string `json:"to"`
	Via   string `json:"via"`
	Build string `json:"build"`
}

// Message passed from the client when declaring who to support with their support order.
type SupportMessage struct {
	BaseMessage
	From   string `json:"from"`   // Name of the area in which the support order is placed.
	To     string `json:"to"`     // Name of the area to which the support order is going.
	Player string `json:"player"` // ID of the player in the destination area to support.
}

// Message passed from the client when they want to quit the game.
type QuitMessage struct {
	BaseMessage
}

// Message passed from the client when they vote to kick another player.
type KickMessage struct {
	BaseMessage
	Player string `json:"player"` // ID of the player to votekick.
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
