package messages

import (
	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/lobby"
)

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
	lobby.Message // Type: MessageSubmitOrders.

	// List of submitted orders.
	Orders []board.Order `json:"orders"`
}

// Message sent from client when declaring who to support with their support order.
// Forwarded by server to all clients to show who were given support.
type GiveSupport struct {
	lobby.Message // Type: MessageGiveSupport.

	// Name of the area in which the support order is placed.
	From string `json:"from"`

	// Name of the area to which the support order is going.
	To string `json:"to"`

	// ID of the player in the destination area to support.
	Player string `json:"player"`
}

// Message sent from client when they want to quit the game.
type Quit struct {
	lobby.Message // Type: MessageQuit.
}

// Message sent from client when they vote to kick another player.
type Kick struct {
	lobby.Message // Type: MessageKick.

	// ID of the player to votekick.
	Player string `json:"player"`
}

// Message passed from the client during winter council voting.
// Used for the throne expansion.
type WinterVote struct {
	lobby.Message // Type: MessageWinterVote

	// ID of the player that the submitting player votes for.
	Player string `json:"player"`
}

// Message passed from the client with the sword to declare where they want to use it.
// Used for the throne expansion.
type Sword struct {
	lobby.Message // Type: MessageSword.

	// Name of the area in which the player wants to use the sword in battle.
	Area string `json:"area"`

	// Index of the battle in which to use the sword, in case of several battles in the area.
	BattleIndex int `json:"battleIndex"`
}

// Message passed from the client with the raven when they want to spy on another player's orders.
// Used for the throne expansion.
type Raven struct {
	lobby.Message // Type: MessageRaven.

	// ID of the player on whom to spy.
	Player string `json:"player"`
}
