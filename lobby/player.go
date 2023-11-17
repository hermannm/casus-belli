package lobby

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/devlog/log"
)

// A player connected to a game lobby.
type Player struct {
	username Username
	// Must hold lock to access safely.
	socket *websocket.Conn
	// Blank until selected. Must hold lock to access safely before the game has started.
	gameFaction game.PlayerFaction
	// Must hold lock to access safely.
	readyToStartGame bool
	lock             sync.RWMutex
	log              log.Logger
}

type Username string

func newPlayer(username Username, socket *websocket.Conn, lobbyLogger log.Logger) *Player {
	return &Player{
		username:         username,
		socket:           socket,
		gameFaction:      "",
		readyToStartGame: false,
		lock:             sync.RWMutex{},
		log:              lobbyLogger.With(slog.Any("player", username)),
	}
}

func (player *Player) selectFaction(faction game.PlayerFaction, lobby *Lobby) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	if !slices.Contains(lobby.game.Factions, faction) {
		return fmt.Errorf("requested faction '%s' is invalid", faction)
	}

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

	player.lock.Lock()
	defer player.lock.Unlock()

	player.gameFaction = faction
	return nil
}

func (player *Player) setReadyToStartGame(ready bool) error {
	player.lock.Lock()
	defer player.lock.Unlock()

	if ready && player.gameFaction == "" {
		return errors.New("must select game faction before setting ready status")
	}

	player.readyToStartGame = ready
	return nil
}
