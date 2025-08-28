package lobby

import (
	"fmt"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"hermannm.dev/devlog/log"

	"hermannm.dev/casus-belli/server/game"
)

// A player connected to a game lobby.
type Player struct {
	username Username
	// Must hold lock to access safely.
	socket *websocket.Conn
	// Blank until selected. Must hold lock to access safely before the game has started.
	gameFaction game.PlayerFaction
	lock        sync.RWMutex
	log         log.Logger
}

type Username string

func newPlayer(username Username, socket *websocket.Conn, lobbyLogger log.Logger) *Player {
	return &Player{
		username:    username,
		socket:      socket,
		gameFaction: "",
		lock:        sync.RWMutex{},
		log:         lobbyLogger.With("player", username),
	}
}

func (player *Player) selectFaction(faction game.PlayerFaction, lobby *Lobby) error {
	if faction != "" {
		if !slices.Contains(lobby.game.PlayerFactions, faction) {
			return fmt.Errorf("requested faction '%s' is invalid", faction)
		}

		lobby.lock.RLock()
		defer lobby.lock.RUnlock()

		var takenBy Username
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
	}

	player.lock.Lock()
	defer player.lock.Unlock()

	player.gameFaction = faction
	return nil
}
