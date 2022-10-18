package lobby

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// A player connected to a game lobby.
type Player struct {
	lock   sync.RWMutex
	socket *websocket.Conn

	username string

	gameID          string              // Blank until selected. Must be selected before starting the game.
	gameMsgReceiver GameMessageReceiver // nil until gameID is set.

	ready bool // Starts as false, must be true before starting the game.
}

// Returns the player's username, with the player's game ID if it is set.
func (player *Player) String() string {
	player.lock.RLock()
	defer player.lock.RUnlock()

	if player.gameID == "" {
		return player.username
	} else {
		return fmt.Sprintf("%s (game ID %s)", player.username, player.gameID)
	}
}

// Marshals the given message to JSON and sends it over the player's socket connection.
// Returns an error if the player is inactive, or if the marshaling/sending failed.
func (player *Player) send(msg any) error {
	player.lock.Lock()

	if err := player.socket.WriteJSON(msg); err != nil {
		player.lock.Unlock()
		return fmt.Errorf("failed to send message to player %s: %w", player.String(), err)
	}

	player.lock.Unlock()
	return nil
}

// Attempts to select the given game ID for the player in the lobby.
func (player *Player) selectGameID(gameID string, lobby *Lobby) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	validGameID := false
	for _, id := range lobby.game.PlayerIDs() {
		if id == gameID {
			validGameID = true
			break
		}
	}
	if !validGameID {
		return fmt.Errorf("requested game ID \"%s\" is invalid", gameID)
	}

	gameIDTakenBy := ""
	for _, player := range lobby.players {
		player.lock.RLock()
		if player.gameID == gameID {
			gameIDTakenBy = player.username
		}
		player.lock.RUnlock()
	}
	if gameIDTakenBy != "" {
		return fmt.Errorf("requested game ID \"%s\" already taken by %s", gameID, gameIDTakenBy)
	}

	player.lock.Lock()
	defer player.lock.Unlock()

	player.gameID = gameID
	return nil
}
