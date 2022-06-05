package messages

import (
	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/gameserver"
)

// Messages from server to client.
const (
	MessageSupportRequest     = "supportRequest"
	MessageOrdersReceived     = "ordersReceived"
	MessageOrdersConfirmation = "ordersConfirmation"
	MessageBattleResult       = "battleResult"
	MessageWinner             = "winner"
)

// Message sent from server when asking a supporting player who to support in an embattled area.
type SupportRequest struct {
	gameserver.Message // Type: MessageSupportRequest.

	// The area from which support is asked, where the asked player should have a support order.
	SupportingArea string `json:"supportingArea"`

	// The embattled area that the player is asked to support.
	EmbattledArea string `json:"embattledArea"`

	// List of possible players to support in the battle.
	Battlers []string `json:"battlers"`
}

// Message sent from server to all clients when valid orders are received from all players.
type OrdersReceived struct {
	gameserver.Message // Type: MessageOrdersReceived.

	// Maps a player's ID to their submitted orders.
	Orders map[string][]board.Order `json:"orders"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmation struct {
	gameserver.Message // Type: MessageOrdersConfirmation.

	// The player who submitted orders.
	Player string `json:"player"`
}

// Message sent from server to all clients when a battle result is calculated.
type BattleResult struct {
	gameserver.Message // Type: MessageBattleResult.

	// The relevant battle result.
	Battle board.Battle `json:"battle"`
}

// Message sent from server to all clients when the game is won.
type Winner struct {
	gameserver.Message // Type: MessageWinner

	// Player tag of the game's winner.
	Winner string `json:"winner"`
}

func SendSupportRequest(
	to gameserver.Sendable, supportingArea string, embattledArea string, battlers []string,
) {
	to.Send(SupportRequest{
		Message:        gameserver.Message{Type: MessageBattleResult},
		SupportingArea: supportingArea,
		EmbattledArea:  embattledArea,
		Battlers:       battlers,
	})
}

func SendOrdersReceived(to gameserver.Sendable, orders map[string][]board.Order) {
	to.Send(OrdersReceived{
		Message: gameserver.Message{Type: MessageOrdersReceived}, Orders: orders,
	})
}

func SendOrdersConfirmation(to gameserver.Sendable, player string) {
	to.Send(OrdersConfirmation{
		Message: gameserver.Message{Type: MessageOrdersConfirmation}, Player: player,
	})
}

func SendBattleResult(to gameserver.Sendable, battle board.Battle) {
	to.Send(BattleResult{Message: gameserver.Message{Type: MessageBattleResult}, Battle: battle})
}

func SendWinner(to gameserver.Sendable, winner string) {
	to.Send(Winner{Message: gameserver.Message{Type: MessageWinner}, Winner: winner})
}
