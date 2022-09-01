package messages

import (
	"hermannm.dev/bfh-server/game/board"
)

type message map[string]any

const errorMsgID = "error"

// Message sent from server when an error occurs.
type errorMsg struct {
	Error string `json:"error"`
}

const supportRequestMsgID = "supportRequest"

// Message sent from server when asking a supporting player who to support in an embattled area.
type supportRequestMsg struct {
	// The area from which support is asked, where the asked player should have a support order.
	SupportingArea string `json:"supportingArea"`

	// List of possible players to support in the battle.
	Battlers []string `json:"battlers"`
}

const orderRequestMsgID = "orderRequest"

type orderRequestMsg struct{}

const ordersReceivedMsgID = "ordersReceived"

// Message sent from server to all clients when valid orders are received from all players.
type ordersReceivedMsg struct {
	// Maps a player's ID to their submitted orders.
	Orders map[string][]board.Order `json:"orders"`
}

const ordersConfirmationMsgID = "ordersConfirmation"

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type ordersConfirmationMsg struct {
	// The player who submitted orders.
	Player string `json:"player"`
}

const battleResultsMsgID = "battleResults"

// Message sent from server to all clients when a battle result is calculated.
type battleResultsMsg struct {
	// The relevant battle result.
	Battles []board.Battle `json:"battles"`
}

const winnerMsgID = "winner"

// Message sent from server to all clients when the game is won.
type winnerMsg struct {
	// Player tag of the game's winner.
	Winner string `json:"winner"`
}

const submitOrdersMsgID = "submitOrders"

// Message sent from client when submitting orders.
type submitOrdersMsg struct {
	// List of submitted orders.
	Orders []board.Order `json:"orders"`
}

const giveSupportMsgID = "giveSupport"

// Message sent from client when declaring who to support with their support order.
// Forwarded by server to all clients to show who were given support.
type giveSupportMsg struct {
	// Name of the area in which the support order is placed.
	From string `json:"from"`

	// ID of the player in the destination area to support.
	Player string `json:"player"`
}

const winterVoteMsgID = "winterVote"

// Message passed from the client during winter council voting.
// Used for the throne expansion.
type winterVoteMsg struct {
	// ID of the player that the submitting player votes for.
	Player string `json:"player"`
}

const swordMsgID = "sword"

// Message passed from the client with the swordMsg to declare where they want to use it.
// Used for the throne expansion.
type swordMsg struct {
	// Name of the area in which the player wants to use the sword in battle.
	Area string `json:"area"`

	// Index of the battle in which to use the sword, in case of several battles in the area.
	BattleIndex int `json:"battleIndex"`
}

const ravenMsgID = "raven"

// Message passed from the client with the ravenMsg when they want to spy on another player's orders.
// Used for the throne expansion.
type ravenMsg struct {
	// ID of the player on whom to spy.
	Player string `json:"player"`
}
