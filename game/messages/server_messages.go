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
