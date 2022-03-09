package messages

// Messages from client to server.
const (
	MessageSubmitOrders = "submitOrders"
	MessageGiveSupport  = "giveSupport"
	MessageQuit         = "quit"
	MessageKick         = "kick"
)

// Client messages used for the throne expansion.
const (
	MessageWinterVote = "winterVote"
	MessageSword      = "sword"
	MessageRaven      = "raven"
)

// Message sent from client when submitting orders.
type SubmitOrders struct {
	Base
	Orders []Order `json:"orders"`
}

// Message sent from client when declaring who to support with their support order.
// Forwarded by server to all clients to show who were given support.
type GiveSupport struct {
	Base
	From   string `json:"from"`   // Name of the area in which the support order is placed.
	To     string `json:"to"`     // Name of the area to which the support order is going.
	Player string `json:"player"` // ID of the player in the destination area to support.
}

// Message sent from client when they want to quit the game.
type Quit struct {
	Base
}

// Message sent from client when they vote to kick another player.
type Kick struct {
	Base
	Player string `json:"player"` // ID of the player to votekick.
}

// Message passed from the client during winter council voting.
// Used for the throne expansion.
type WinterVote struct {
	Base
	Player string `json:"player"` // ID of the player that the submitting player votes for.
}

// Message passed from the client with the sword to declare where they want to use it.
// Used for the throne expansion.
type Sword struct {
	Base
	Area        string // Name of the area in which the player wants to use the sword in battle.
	BattleIndex int    // Index of the battle in which to use the sword, in case of several battles in the area.
}

// Message passed from the client with the raven when they want to spy on another player's orders.
// Used for the throne expansion.
type Raven struct {
	Base
	Player string // ID of the player on whom to spy.
}
