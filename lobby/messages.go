package lobby

import (
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/enumnames"
)

type Message struct {
	Tag  MessageTag
	Data any
}

// Message sent from server when an error occurs.
type ErrorMessage struct {
	Error string
}

// Message sent from server to all clients when a player's status changes.
type PlayerStatusMessage struct {
	Username         Username
	SelectedFaction  *game.PlayerFaction `json:",omitempty"`
	ReadyToStartGame bool
}

// Message sent to a player when they join a lobby, to inform them about the game and other players.
type LobbyJoinedMessage struct {
	SelectableFactions []game.PlayerFaction
	PlayerStatuses     []PlayerStatusMessage
}

// Message sent from client when they want to select a faction to play for the game.
type SelectFactionMessage struct {
	Faction game.PlayerFaction
}

// Message sent from client to mark themselves as ready to start the game.
// Requires that faction has been selected.
type ReadyToStartGameMessage struct {
	Ready bool
}

// Message sent from a player when the lobby wants to start the game.
// Requires that all players are ready.
type StartGameMessage struct{}

// Message sent from server when asking a supporting player who to support in an embattled region.
type SupportRequestMessage struct {
	SupportingRegion    game.RegionName
	EmbattledRegion     game.RegionName
	SupportableFactions []game.PlayerFaction
}

// Message sent from server to client to signal that client should submit orders.
type OrderRequestMessage struct{}

// Message sent from server to all clients when valid orders are received from all players.
type OrdersReceivedMessage struct {
	OrdersByFaction map[game.PlayerFaction][]game.Order
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmationMessage struct {
	FactionThatSubmittedOrders game.PlayerFaction
}

// Message sent from server to all clients when a battle result is calculated.
type BattleResultsMessage struct {
	Battles []game.Battle
}

// Message sent from server to all clients when the game is won.
type WinnerMessage struct {
	WinningFaction game.PlayerFaction
}

// Message sent from client when submitting orders.
type SubmitOrdersMessage struct {
	Orders []game.Order
}

// Message sent from client when declaring who to support with their support order.
// Forwarded by server to all clients to show who were given support.
type GiveSupportMessage struct {
	SupportingRegion game.RegionName
	EmbattledRegion  game.RegionName

	// Nil if none were supported.
	SupportedFaction *game.PlayerFaction
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
})

func (tag MessageTag) String() string {
	return messageTags.GetNameOrFallback(tag, "INVALID")
}
