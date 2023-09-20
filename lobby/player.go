package lobby

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// A player connected to a game lobby.
type Player struct {
	username            string
	gameMessageReceiver GameMessageReceiver
	socket              *websocket.Conn // Must hold lock to access safely.
	gameID              string          // Must hold lock to access safely. Blank until selected.
	readyToStartGame    bool            // Must hold lock to access safely.
	lock                sync.RWMutex
}

func newPlayer(username string, socket *websocket.Conn) (*Player, error) {
	if username == "" {
		return nil, errors.New("player cannot have blank username")
	}

	return &Player{
		username:            username,
		gameMessageReceiver: newGameMessageReceiver(),
		lock:                sync.RWMutex{},
		socket:              socket,
		gameID:              "",
		readyToStartGame:    false,
	}, nil
}

func (player *Player) selectGameID(gameID string, lobby *Lobby) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	validGameID := false
	for _, id := range lobby.game.PlayerIDs {
		if id == gameID {
			validGameID = true
			break
		}
	}
	if !validGameID {
		return fmt.Errorf("requested game ID '%s' is invalid", gameID)
	}

	var gameIDTakenBy string
	for _, player := range lobby.players {
		player.lock.RLock()
		if player.gameID == gameID {
			gameIDTakenBy = player.username
		}
		player.lock.RUnlock()
	}
	if gameIDTakenBy != "" {
		return fmt.Errorf(
			"requested game ID '%s' already taken by user '%s'",
			gameID,
			gameIDTakenBy,
		)
	}

	player.lock.Lock()
	defer player.lock.Unlock()

	player.gameID = gameID
	return nil
}

func (player *Player) setReadyToStartGame(ready bool) error {
	player.lock.Lock()
	defer player.lock.Unlock()

	if ready && player.gameID == "" {
		return errors.New("must select game ID before setting ready status")
	}

	player.readyToStartGame = ready
	return nil
}

// Returns the player's username, along with the player's game ID if it is set.
func (player *Player) String() string {
	player.lock.RLock()
	defer player.lock.RUnlock()

	if player.gameID == "" {
		return player.username
	} else {
		return fmt.Sprintf("%s (game ID %s)", player.username, player.gameID)
	}
}
