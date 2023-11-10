package lobby

import (
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/enumnames"
)

type Message struct {
	Tag  MessageTag `json:"tag"`
	Data any        `json:"data"`
}

// Message sent from server when an error occurs.
type ErrorMessage struct {
	Error string `json:"error"`
}

// Message sent from server to all clients when a player's status changes.
type PlayerStatusMessage struct {
	Username         string                   `json:"username"`
	SelectedFaction  *gametypes.PlayerFaction `json:"selectedFaction,omitempty"`
	ReadyToStartGame bool                     `json:"readyToStartGame"`
}

// Message sent to a player when they join a lobby, to inform them about the game and other players.
type LobbyJoinedMessage struct {
	SelectableFactions []gametypes.PlayerFaction `json:"selectableFactions"`
	PlayerStatuses     []PlayerStatusMessage     `json:"playerStatuses"`
}

// Message sent from client when they want to select a faction to play for the game.
type SelectFactionMessage struct {
	Faction gametypes.PlayerFaction `json:"faction"`
}

// Message sent from client to mark themselves as ready to start the game.
// Requires that faction has been selected.
type ReadyToStartGameMessage struct {
	Ready bool `json:"ready"`
}

// Message sent from a player when the lobby wants to start the game.
// Requires that all players are ready.
type StartGameMessage struct{}

// Message sent from server when asking a supporting player who to support in an embattled region.
type SupportRequestMessage struct {
	SupportingRegion    string                    `json:"supportingRegion"`
	EmbattledRegion     string                    `json:"embattledRegion"`
	SupportableFactions []gametypes.PlayerFaction `json:"supportableFactions"`
}

// Message sent from server to client to signal that client should submit orders.
type OrderRequestMessage struct{}

// Message sent from server to all clients when valid orders are received from all players.
type OrdersReceivedMessage struct {
	OrdersByFaction map[gametypes.PlayerFaction][]gametypes.Order `json:"ordersByFaction"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmationMessage struct {
	FactionThatSubmittedOrders gametypes.PlayerFaction `json:"factionThatSubmittedOrders"`
}

// Message sent from server to all clients when a battle result is calculated.
type BattleResultsMessage struct {
	Battles []gametypes.Battle `json:"battles"`
}

// Message sent from server to all clients when the game is won.
type WinnerMessage struct {
	WinningFaction gametypes.PlayerFaction `json:"winningFaction"`
}

// Message sent from client when submitting orders.
type SubmitOrdersMessage struct {
	Orders []gametypes.Order `json:"orders"`
}

// Message sent from client when declaring who to support with their support order.
// Forwarded by server to all clients to show who were given support.
type GiveSupportMessage struct {
	SupportingRegion string `json:"supportingRegion"`
	EmbattledRegion  string `json:"embattledRegion"`

	// Nil if none were supported.
	SupportedFaction *gametypes.PlayerFaction `json:"supportedFaction"`
}

// Message sent from the client during winter council voting.
// Used for the throne expansion.
type WinterVoteMessage struct {
	FactionVotedFor gametypes.PlayerFaction `json:"factionVotedFor"`
}

// Message passed from the client with the Sword to declare where they want to use it.
// Used for the throne expansion.
type SwordMessage struct {
	Region string `json:"region"`

	// Index of the battle in which to use the sword, in case of several battles in the region.
	BattleIndex int `json:"battleIndex"`
}

// Message sent from the client with the Raven when they want to spy on another player's
// orders.
// Used for the throne expansion.
type RavenMessage struct {
	FactionToSpyOn string `json:"factionToSpyOn"`
}

type MessageTag uint8

const (
	MessageTagError MessageTag = iota + 1
	MessageTagPlayerStatus
	MessageTagLobbyJoined
	MessageTagSelectFaction
	MessageTagReady
	MessageTagStartGame
	MessageTagSupportRequest
	MessageTagOrderRequest
	MessageTagOrdersReceived
	MessageTagOrdersConfirmation
	MessageTagBattleResults
	MessageTagWinner
	MessageTagSubmitOrders
	MessageTagGiveSupport
	MessageTagWinterVote
	MessageTagSword
	MessageTagRaven
)

var messageTags = enumnames.NewMap(map[MessageTag]string{
	MessageTagError:              "Error",
	MessageTagPlayerStatus:       "PlayerStatus",
	MessageTagLobbyJoined:        "LobbyJoined",
	MessageTagSelectFaction:      "SelectFaction",
	MessageTagReady:              "Ready",
	MessageTagStartGame:          "StartGame",
	MessageTagSupportRequest:     "SupportRequest",
	MessageTagOrderRequest:       "OrderRequest",
	MessageTagOrdersReceived:     "OrdersReceived",
	MessageTagOrdersConfirmation: "OrdersConfirmation",
	MessageTagBattleResults:      "BattleResults",
	MessageTagWinner:             "Winner",
	MessageTagSubmitOrders:       "SubmitOrders",
	MessageTagGiveSupport:        "GiveSupport",
	MessageTagWinterVote:         "WinterVote",
	MessageTagSword:              "Sword",
	MessageTagRaven:              "Raven",
})

func (tag MessageTag) String() string {
	return messageTags.GetNameOrFallback(tag, "INVALID")
}
