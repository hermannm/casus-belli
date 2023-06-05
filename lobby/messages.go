package lobby

import "hermannm.dev/bfh-server/game/gametypes"

// Messages map a single key, the message type, to an object determined by its type.
type Message map[MessageType]any

func (message Message) Type() MessageType {
	for messageType := range message {
		return messageType
	}
	return "" // Messages should always have its type set, so this should never be reached
}

type MessageType string

const (
	messageTypeError              MessageType = "error"
	messageTypePlayerStatus       MessageType = "playerStatus"
	messageTypeLobbyJoined        MessageType = "lobbyJoined"
	messageTypeSelectGameID       MessageType = "selectGameId"
	messageTypeReady              MessageType = "ready"
	messageTypeStartGame          MessageType = "startGame"
	messageTypeSupportRequest     MessageType = "supportRequest"
	messageTypeOrderRequest       MessageType = "orderRequest"
	messageTypeOrdersReceived     MessageType = "ordersReceived"
	messageTypeOrdersConfirmation MessageType = "ordersConfirmation"
	messageTypeBattleResults      MessageType = "battleResults"
	messageTypeWinner             MessageType = "winner"
	messageTypeSubmitOrders       MessageType = "submitOrders"
	messageTypeGiveSupport        MessageType = "giveSupport"
	messageTypeWinterVote         MessageType = "winterVote"
	messageTypeSword              MessageType = "sword"
	messageTypeRaven              MessageType = "raven"
)

// Message sent from server when an error occurs.
type ErrorMessage struct {
	Error string `json:"error"`
}

// Message sent from server to all clients when a player's status changes.
type PlayerStatusMessage struct {
	Username         string  `json:"username"`
	GameID           *string `json:"gameId,omitempty"`
	ReadyToStartGame bool    `json:"ready"`
}

// Message sent to a player when they join a lobby, to inform them about the game and other players.
type LobbyJoinedMessage struct {
	SelectableGameIDs []string              `json:"selectableGameIds"`
	PlayerStatuses    []PlayerStatusMessage `json:"playerStatuses"`
}

// Message sent from client when they want to select a game ID.
type SelectGameIDMessage struct {
	GameID string `json:"gameId"`
}

// Message sent from client to mark themselves as ready to start the game.
// Requires game ID being selected.
type ReadyToStartGameMessage struct {
	Ready bool `json:"ready"`
}

// Message sent from a player when the lobby wants to start the game.
// Requires that all players are ready.
type StartGameMessage struct{}

// Message sent from server when asking a supporting player who to support in an embattled region.
type SupportRequestMessage struct {
	SupportingRegion   string   `json:"supportingRegion"`
	EmbattledRegion    string   `json:"embattledRegion"`
	SupportablePlayers []string `json:"supportablePlayers"`
}

// Message sent from server to client to signal that client should submit orders.
type OrderRequestMessage struct{}

// Message sent from server to all clients when valid orders are received from all players.
type OrdersReceivedMessage struct {
	// Maps a player's ID to their submitted orders.
	PlayerOrders map[string][]gametypes.Order `json:"playerOrders"`
}

// Message sent from server to all clients when valid orders are received from a player.
// Used to show who the server is waiting for.
type OrdersConfirmationMessage struct {
	PlayerWhoSubmittedOrders string `json:"playerWhoSubmittedOrders"`
}

// Message sent from server to all clients when a battle result is calculated.
type BattleResultsMessage struct {
	Battles []gametypes.Battle `json:"battles"`
}

// Message sent from server to all clients when the game is won.
type WinnerMessage struct {
	Winner string `json:"winner"`
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
	SupportedPlayer *string `json:"supportedPlayer"`
}

// Message passed from the client during winter council voting.
// Used for the throne expansion.
type WinterVoteMessage struct {
	// ID of the player that the submitting player votes for.
	Player string `json:"player"`
}

// Message passed from the client with the SwordMessage to declare where they want to use it.
// Used for the throne expansion.
type SwordMessage struct {
	// Name of the region in which the player wants to use the sword in battle.
	Region string `json:"region"`

	// Index of the battle in which to use the sword, in case of several battles in the region.
	BattleIndex int `json:"battleIndex"`
}

// Message passed from the client with the RavenMessage when they want to spy on another player's
// orders.
// Used for the throne expansion.
type RavenMessage struct {
	// ID of the player on whom to spy.
	Player string `json:"player"`
}
