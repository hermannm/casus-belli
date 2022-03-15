package messages

import (
	"hermannm.dev/bfh-server/game/board"
)

// Messages from server to client.
const (
	MessageAskSupport         = "askSupport"
	MessageOrdersReceived     = "ordersReceived"
	MessageOrdersConfirmation = "ordersConfirmation"
	MessageBattleResult       = "battleResult"
)

// Message sent from server when asking a supporting player who to support in an embattled area.
type AskSupport struct {
	Base // Type: MessageAskSupport.

	// The area from which support is asked, where the asked player should have a support order.
	From string `json:"from"`

	// The embattled area that the player is asked to support.
	To string `json:"to"`

	// List of possible players to support in the battle.
	Battlers []string `json:"battlers"`
}

// Message sent from server to all clients when valid orders are received from all players.
type OrdersReceived struct {
	Base // Type: MessageOrdersReceived.

	// Maps a player's ID to their submitted orders.
	Orders map[string][]board.Order `json:"orders"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmation struct {
	Base // Type: MessageOrdersConfirmation.

	// The player who submitted orders.
	Player string `json:"player"`
}

// Message sent from server to all clients when a battle result is calculated.
type BattleResult struct {
	Base // Type: MessageBattleResult.

	// The relevant battle result.
	Battle board.Battle `json:"battle"`
}
