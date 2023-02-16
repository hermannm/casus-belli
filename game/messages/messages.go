package messages

import (
	"hermannm.dev/bfh-server/game/board"
)

// Messages map a single key, the message ID, to an object determined by the message ID.
type message map[string]any

// IDs for game-specific messages.
const (
	errorMsgID              = "error"
	supportRequestMsgID     = "supportRequest"
	orderRequestMsgID       = "orderRequest"
	ordersReceivedMsgID     = "ordersReceived"
	ordersConfirmationMsgID = "ordersConfirmation"
	battleResultsMsgID      = "battleResults"
	winnerMsgID             = "winner"
	submitOrdersMsgID       = "submitOrders"
	giveSupportMsgID        = "giveSupport"
	winterVoteMsgID         = "winterVote"
	swordMsgID              = "sword"
	ravenMsgID              = "raven"
)

// Message sent from server when an error occurs.
type errorMsg struct {
	Error string `json:"error"`
}

// Message sent from server when asking a supporting player who to support in an embattled region.
type supportRequestMsg struct {
	// The region from which support is asked, where the asked player should have a support order.
	SupportingRegion string `json:"supportingRegion"`

	// List of possible players to support in the battle.
	SupportablePlayers []string `json:"supportablePlayers"`
}

// Message sent from server to client to signal that client should submit orders.
type orderRequestMsg struct{}

// Message sent from server to all clients when valid orders are received from all players.
type ordersReceivedMsg struct {
	// Maps a player's ID to their submitted orders.
	PlayerOrders map[string][]board.Order `json:"playerOrders"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type ordersConfirmationMsg struct {
	// The player who submitted orders.
	Player string `json:"player"`
}

// Message sent from server to all clients when a battle result is calculated.
type battleResultsMsg struct {
	// The relevant battle result.
	Battles []board.Battle `json:"battles"`
}

// Message sent from server to all clients when the game is won.
type winnerMsg struct {
	// Player tag of the game's winner.
	Winner string `json:"winner"`
}

// Message sent from client when submitting orders.
type submitOrdersMsg struct {
	// List of submitted orders.
	Orders []board.Order `json:"orders"`
}

// Message sent from client when declaring who to support with their support order.
// Forwarded by server to all clients to show who were given support.
type giveSupportMsg struct {
	// Name of the region where the supporting player has their support order.
	SupportingRegion string `json:"supportingRegion"`

	// ID of the player in the destination region to support.
	// Nil if none were supported.
	SupportedPlayer *string `json:"supportedPlayer"`
}

// Message passed from the client during winter council voting.
// Used for the throne expansion.
type winterVoteMsg struct {
	// ID of the player that the submitting player votes for.
	Player string `json:"player"`
}

// Message passed from the client with the swordMsg to declare where they want to use it.
// Used for the throne expansion.
type swordMsg struct {
	// Name of the region in which the player wants to use the sword in battle.
	Region string `json:"region"`

	// Index of the battle in which to use the sword, in case of several battles in the region.
	BattleIndex int `json:"battleIndex"`
}

// Message passed from the client with the ravenMsg when they want to spy on another player's
// orders.
// Used for the throne expansion.
type ravenMsg struct {
	// ID of the player on whom to spy.
	Player string `json:"player"`
}
