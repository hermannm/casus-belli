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
	Username string  `json:"username"`
	GameID   *string `json:"gameID"`
	Ready    bool    `json:"ready"`
}

// Message sent to a player when they join a lobby, to inform them about other players.
type lobbyJoinedMsg struct {
	PlayerStatuses []playerStatusMsg `json:"playerStatuses"`
}

// Message sent from client when they want to select a game ID.
type selectGameIDMsg struct {
	GameID string `json:"gameID"`
}

// Message sent from client to mark themselves as ready to start the game (requires game ID being selected).
type readyMsg struct {
	Ready bool `json:"ready"`
}

// Message sent from a player when the lobby wants to start the game (requires that all players are ready).
type startGameMsg struct{}
