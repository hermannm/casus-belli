package messages

import (
	"hermannm.dev/bfh-server/game/board"
)

// Messages from server to client.
const (
	msgError              = "error"
	msgSupportRequest     = "supportRequest"
	msgOrderRequest       = "orderRequest"
	msgOrdersReceived     = "ordersReceived"
	msgOrdersConfirmation = "ordersConfirmation"
	msgBattleResult       = "battleResult"
	msgWinner             = "winner"
)

// Messages from client to server.
const (
	msgSubmitOrders = "submitOrders"
	msgGiveSupport  = "giveSupport"
)

// Client messages used for the throne expansion.
const (
	msgWinterVote = "winterVote"
	msgSword      = "sword"
	msgRaven      = "raven"
)

// Message sent from server when an error occurs.
type errorMsg struct {
	Type  string `json:"type"` // msgError
	Error string `json:"error"`
}

// Message sent from server when asking a supporting player who to support in an embattled area.
type supportRequestMsg struct {
	Type string `json:"type"` // msgSupportRequest

	// The area from which support is asked, where the asked player should have a support order.
	SupportingArea string `json:"supportingArea"`

	// List of possible players to support in the battle.
	Battlers []string `json:"battlers"`
}

type orderRequestMsg struct {
	Type string `json:"type"`
}

// Message sent from server to all clients when valid orders are received from all players.
type ordersReceivedMsg struct {
	Type string `json:"type"` // msgOrdersReceived

	// Maps a player's ID to their submitted orders.
	Orders map[string][]board.Order `json:"orders"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type ordersConfirmationMsg struct {
	Type string `json:"type"` // msgOrdersConfirmation

	// The player who submitted orders.
	Player string `json:"player"`
}

// Message sent from server to all clients when a battle result is calculated.
type battleResultMsg struct {
	Type string `json:"type"` // msgBattleResult

	// The relevant battle result.
	Battle board.Battle `json:"battle"`
}

// Message sent from server to all clients when the game is won.
type winnerMsg struct {
	Type string `json:"type"` // msgWinner

	// Player tag of the game's winner.
	Winner string `json:"winner"`
}

// Message sent from client when submitting orders.
type submitOrdersMsg struct {
	Type string `json:"type"` // msgSubmitOrders

	// List of submitted orders.
	Orders []board.Order `json:"orders"`
}

// Message sent from client when declaring who to support with their support order.
// Forwarded by server to all clients to show who were given support.
type giveSupportMsg struct {
	Type string `json:"type"` // msgGiveSupport

	// Name of the area in which the support order is placed.
	From string `json:"from"`

	// ID of the player in the destination area to support.
	Player string `json:"player"`
}

// Message passed from the client during winter council voting.
// Used for the throne expansion.
type winterVoteMsg struct {
	Type string `json:"type"` // msgWinterVote

	// ID of the player that the submitting player votes for.
	Player string `json:"player"`
}

// Message passed from the client with the swordMsg to declare where they want to use it.
// Used for the throne expansion.
type swordMsg struct {
	Type string `json:"type"` // msgSword

	// Name of the area in which the player wants to use the sword in battle.
	Area string `json:"area"`

	// Index of the battle in which to use the sword, in case of several battles in the area.
	BattleIndex int `json:"battleIndex"`
}

// Message passed from the client with the ravenMsg when they want to spy on another player's orders.
// Used for the throne expansion.
type ravenMsg struct {
	Type string `json:"type"` // msgRaven

	// ID of the player on whom to spy.
	Player string `json:"player"`
}
