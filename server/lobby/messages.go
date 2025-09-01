package lobby

import (
	"log/slog"

	"hermannm.dev/enumnames"

	"hermannm.dev/casus-belli/server/game"
)

type Message struct {
	Tag  MessageTag `json:"Tag"`
	Data any        `json:"Data"`
}

// Implements [slog.LogValuer].
func (message Message) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("tag", message.Tag.String()),
		slog.Any("data", message.Data),
	)
}

type ReceivedMessage struct {
	Tag          MessageTag         `json:"Tag"`
	Data         any                `json:"Data"`
	ReceivedFrom game.PlayerFaction `json:"ReceivedFrom"`
}

// Message sent from server when an error occurs.
type ErrorMessage struct {
	Error string `json:"Error"`
}

// Message sent to a player when they join a lobby, to inform them about the game and other players.
type LobbyJoinedMessage struct {
	SelectableFactions []game.PlayerFaction  `json:"SelectableFactions"`
	PlayerStatuses     []PlayerStatusMessage `json:"PlayerStatuses"`
}

// Message sent from server to all clients when a player's status changes.
type PlayerStatusMessage struct {
	Username        Username           `json:"Username"`
	SelectedFaction game.PlayerFaction `json:"SelectedFaction,omitempty"`
}

// Message sent from client when they want to select a faction to play for the game.
type SelectFactionMessage struct {
	Faction game.PlayerFaction `json:"Faction"`
}

// Message sent from a player when the lobby wants to start the game.
// Requires that all players have selected a faction.
type StartGameMessage struct{}

// Message sent from server when the game starts.
type GameStartedMessage struct {
	Board game.Board `json:"Board"`
}

// Message sent from server to client to signal that client should submit orders.
type OrderRequestMessage struct {
	Season game.Season `json:"Season"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmationMessage struct {
	FactionThatSubmittedOrders game.PlayerFaction `json:"FactionThatSubmittedOrders"`
}

// Message sent from server to all clients when valid orders are received from all players.
type OrdersReceivedMessage struct {
	// All orders will be non-nil.
	OrdersByFaction map[game.PlayerFaction][]*game.Order `json:"OrdersByFaction"`
}

// Message sent from server to all clients when a battle has begun.
type BattleAnnouncementMessage struct {
	Battle game.Battle `json:"Battle"`
}

// Message sent from server to all clients when a battle has finished resolving.
type BattleResultsMessage struct {
	Battle game.Battle `json:"Battle"`
}

// Message sent from server to all clients when the game is won.
type WinnerMessage struct {
	WinningFaction game.PlayerFaction `json:"WinningFaction"`
}

// Message sent from client when submitting orders.
type SubmitOrdersMessage struct {
	// All elements must be non-nil (checked in [Lobby.AwaitOrders]).
	Orders []*game.Order `json:"Orders"`
}

// Message sent from client when they roll the dice in a battle.
type DiceRollMessage struct{}

// Message sent from client when declaring who to support with their support order.
type GiveSupportMessage struct {
	EmbattledRegion game.RegionName `json:"EmbattledRegion"`

	// Blank if none were supported.
	SupportedFaction game.PlayerFaction `json:"SupportedFaction,omitempty"`
}

type MessageTag uint8

const (
	MessageTagError MessageTag = iota + 1
	MessageTagLobbyJoined
	MessageTagPlayerStatus
	MessageTagSelectFaction
	MessageTagStartGame
	MessageTagGameStarted
	MessageTagOrderRequest
	MessageTagOrdersReceived
	MessageTagOrdersConfirmation
	MessageTagBattleAnnouncement
	MessageTagBattleResults
	MessageTagWinner
	MessageTagSubmitOrders
	MessageTagDiceRoll
	MessageTagGiveSupport
)

var messageTags = enumnames.NewMap(
	map[MessageTag]string{
		MessageTagError:              "Error",
		MessageTagLobbyJoined:        "LobbyJoined",
		MessageTagPlayerStatus:       "PlayerStatus",
		MessageTagSelectFaction:      "SelectFaction",
		MessageTagStartGame:          "StartGame",
		MessageTagGameStarted:        "GameStarted",
		MessageTagOrderRequest:       "OrderRequest",
		MessageTagOrdersConfirmation: "OrdersConfirmation",
		MessageTagOrdersReceived:     "OrdersReceived",
		MessageTagBattleAnnouncement: "BattleAnnouncement",
		MessageTagBattleResults:      "BattleResults",
		MessageTagWinner:             "Winner",
		MessageTagSubmitOrders:       "SubmitOrders",
		MessageTagDiceRoll:           "DiceRoll",
		MessageTagGiveSupport:        "GiveSupport",
	},
)

func (tag MessageTag) String() string {
	return messageTags.GetNameOrFallback(tag, "INVALID")
}
