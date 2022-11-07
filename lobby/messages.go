package lobby

// Messages map a single key, the message ID, to an object determined by the message ID.
type message map[string]any

// IDs for lobby-specific messages.
const (
	errorMsgID        = "error"
	playerStatusMsgID = "playerStatus"
	lobbyJoinedMsgID  = "lobbyJoined"
	selectGameIDMsgID = "selectGameId"
	readyMsgID        = "ready"
	startGameMsgID    = "startGame"
)

// Message sent from server when an error occurs.
type errorMsg struct {
	Error string `json:"error"`
}

// Message sent from server to all clients when a player's status changes.
type playerStatusMsg struct {
	// The user's chosen display name.
	Username string `json:"username"`

	// The user's selected game ID. Nil if not selected yet.
	GameID *string `json:"gameID"`

	// Whether the user is ready to start the game.
	Ready bool `json:"ready"`
}

// Message sent to a player when they join a lobby, to inform them about the game and other players.
type lobbyJoinedMsg struct {
	// IDs that the player may select from for this lobby's game.
	// Returns all game IDs, though some may already be taken by other players in the lobby.
	GameIDs []string `json:"gameIDs"`

	// Info about each other player in the lobby.
	PlayerStatuses []playerStatusMsg `json:"playerStatuses"`
}

// Message sent from client when they want to select a game ID.
type selectGameIDMsg struct {
	// The ID that the player wants to select for the game.
	// Will be rejected if already selected by another player.
	GameID string `json:"gameID"`
}

// Message sent from client to mark themselves as ready to start the game.
// Requires game ID being selected.
type readyMsg struct {
	// Whether the player is ready to start the game.
	Ready bool `json:"ready"`
}

// Message sent from a player when the lobby wants to start the game.
// Requires that all players are ready.
type startGameMsg struct{}
