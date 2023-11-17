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
	username         Username
	socket           *websocket.Conn    // Must hold lock to access safely.
	gameFaction      game.PlayerFaction // Must hold lock to access safely. Blank until selected.
	readyToStartGame bool               // Must hold lock to access safely.
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
