package messages

import (
	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/lobby"
)

// Messages from server to client.
const (
	MessageAskSupport         = "askSupport"
	MessageOrdersReceived     = "ordersReceived"
	MessageOrdersConfirmation = "ordersConfirmation"
	MessageBattleResult       = "battleResult"
	MessageWinner             = "winner"
)

// Message sent from server when asking a supporting player who to support in an embattled area.
type AskSupport struct {
	lobby.Message // Type: MessageAskSupport.

	// The area from which support is asked, where the asked player should have a support order.
	From string `json:"from"`

	// The embattled area that the player is asked to support.
	To string `json:"to"`

	// List of possible players to support in the battle.
	Battlers []string `json:"battlers"`
}

// Message sent from server to all clients when valid orders are received from all players.
type OrdersReceived struct {
	lobby.Message // Type: MessageOrdersReceived.

	// Maps a player's ID to their submitted orders.
	Orders map[string][]board.Order `json:"orders"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmation struct {
	lobby.Message // Type: MessageOrdersConfirmation.

	// The player who submitted orders.
	Player string `json:"player"`
}

// Message sent from server to all clients when a battle result is calculated.
type BattleResult struct {
	lobby.Message // Type: MessageBattleResult.

	// The relevant battle result.
	Battle board.Battle `json:"battle"`
}

// Message sent from server to all clients when the game is won.
type Winner struct {
	lobby.Message // Type: MessageWinner

	// Player tag of the game's winner.
	Winner string `json:"winner"`
}
