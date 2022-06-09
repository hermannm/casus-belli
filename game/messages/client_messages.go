package messages

import (
	"hermannm.dev/bfh-server/game/board"
)

// Messages from client to server.
const (
	MsgSubmitOrders = "submitOrders"
	MsgGiveSupport  = "giveSupport"
	MsgQuit         = "quit"
	MsgKick         = "kick"
)

// Client messages used for the throne expansion.
const (
	MsgWinterVote = "winterVote"
	MsgSword      = "sword"
	MsgRaven      = "raven"
)

// Message sent from client when submitting orders.
type SubmitOrders struct {
	Type string `json:"type"` // MsgSubmitOrders

	// List of submitted orders.
	Orders []board.Order `json:"orders"`
}

// Message sent from client when declaring who to support with their support order.
// Forwarded by server to all clients to show who were given support.
type GiveSupport struct {
	Type string `json:"type"` // MsgGiveSupport

	// Name of the area in which the support order is placed.
	From string `json:"from"`

	// Name of the area to which the support order is going.
	To string `json:"to"`

	// ID of the player in the destination area to support.
	Player string `json:"player"`
}

// Message sent from client when they want to quit the game.
type Quit struct {
	Type string `json:"type"` // MsgQuit
}

// Message sent from client when they vote to kick another player.
type Kick struct {
	Type string `json:"type"` // MsgKick

	// ID of the player to votekick.
	Player string `json:"player"`
}

// Message passed from the client during winter council voting.
// Used for the throne expansion.
type WinterVote struct {
	Type string `json:"type"` // MsgWinterVote

	// ID of the player that the submitting player votes for.
	Player string `json:"player"`
}

// Message passed from the client with the sword to declare where they want to use it.
// Used for the throne expansion.
type Sword struct {
	Type string `json:"type"` // MsgSword

	// Name of the area in which the player wants to use the sword in battle.
	Area string `json:"area"`

	// Index of the battle in which to use the sword, in case of several battles in the area.
	BattleIndex int `json:"battleIndex"`
}

// Message passed from the client with the raven when they want to spy on another player's orders.
// Used for the throne expansion.
type Raven struct {
	Type string `json:"type"` // MsgRaven

	// ID of the player on whom to spy.
	Player string `json:"player"`
}
