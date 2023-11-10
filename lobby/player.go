package lobby

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game/gametypes"
)

// A player connected to a game lobby.
type Player struct {
	username            string
	gameMessageReceiver GameMessageReceiver
	socket              *websocket.Conn         // Must hold lock to access safely.
	gameFaction         gametypes.PlayerFaction // Must hold lock to access safely. Blank until selected.
	readyToStartGame    bool                    // Must hold lock to access safely.
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
		gameFaction:         "",
		readyToStartGame:    false,
	}, nil
}

func (player *Player) selectFaction(faction gametypes.PlayerFaction, lobby *Lobby) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	if !slices.Contains(lobby.game.Factions, faction) {
		return fmt.Errorf("requested faction '%s' is invalid", faction)
	}

	var takenBy string
	for _, otherPlayer := range lobby.players {
		if otherPlayer.username == player.username {
			continue
		}

		otherPlayer.lock.RLock()
		if otherPlayer.gameFaction == faction {
			takenBy = otherPlayer.username
		}
		otherPlayer.lock.RUnlock()
	}
	if takenBy != "" {
		return fmt.Errorf(
			"requested faction '%s' already taken by user '%s'",
			faction,
			takenBy,
		)
	}

	player.lock.Lock()
	defer player.lock.Unlock()

	player.gameFaction = faction
	return nil
}

func (player *Player) setReadyToStartGame(ready bool) error {
	player.lock.Lock()
	defer player.lock.Unlock()

	if ready && player.gameFaction == "" {
		return errors.New("must select game ID before setting ready status")
	}

	player.readyToStartGame = ready
	return nil
}

// Returns the player's username, along with the player's game ID if it is set.
func (player *Player) String() string {
	player.lock.RLock()
	defer player.lock.RUnlock()

	if player.gameFaction == "" {
		return fmt.Sprintf("'%s'", player.username)
	} else {
		return fmt.Sprintf("'%s' (%s)", player.username, player.gameFaction)
	}
}
