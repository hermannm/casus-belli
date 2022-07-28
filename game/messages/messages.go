package messages

import (
	"hermannm.dev/bfh-server/game/board"
)

// Messages from server to client.
const (
	MsgError              = "error"
	MsgSupportRequest     = "supportRequest"
	MsgOrderRequest       = "orderRequest"
	MsgOrdersReceived     = "ordersReceived"
	MsgOrdersConfirmation = "ordersConfirmation"
	MsgBattleResult       = "battleResult"
	MsgWinner             = "winner"
)

// Message sent from server when an error occurs.
type Error struct {
	Type  string `json:"type"` // MsgError
	Error string `json:"error"`
}

// Message sent from server when asking a supporting player who to support in an embattled area.
type SupportRequest struct {
	Type string `json:"type"` // MsgSupportRequest

	// The area from which support is asked, where the asked player should have a support order.
	SupportingArea string `json:"supportingArea"`

	// List of possible players to support in the battle.
	Battlers []string `json:"battlers"`
}

type OrderRequest struct {
	Type string `json:"type"`
}

// Message sent from server to all clients when valid orders are received from all players.
type OrdersReceived struct {
	Type string `json:"type"` // MsgOrdersReceived

	// Maps a player's ID to their submitted orders.
	Orders map[string][]board.Order `json:"orders"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmation struct {
	Type string `json:"type"` // MsgOrdersConfirmation

	// The player who submitted orders.
	Player string `json:"player"`
}

// Message sent from server to all clients when a battle result is calculated.
type BattleResult struct {
	Type string `json:"type"` // MsgBattleResult

	// The relevant battle result.
	Battle board.Battle `json:"battle"`
}

// Message sent from server to all clients when the game is won.
type Winner struct {
	Type string `json:"type"` // MsgWinner

	// Player tag of the game's winner.
	Winner string `json:"winner"`
}

// Messages from client to server.
const (
	MsgSubmitOrders = "submitOrders"
	MsgGiveSupport  = "giveSupport"
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

	// ID of the player in the destination area to support.
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
