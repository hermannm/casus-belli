package lobby

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/condqueue"
	"hermannm.dev/devlog/log"
)

// A collection of players for a game.
type Lobby struct {
	name             string
	players          []*Player // Must hold lock to access safely.
	game             *game.Game
	gameStarted      bool // Must hold lock to access safely.
	gameMessageQueue *condqueue.CondQueue[ReceivedMessage]
	registry         *LobbyRegistry
	lock             sync.RWMutex
	log              log.Logger
}

func (lobby *Lobby) getPlayer(faction game.PlayerFaction) (player *Player, foundPlayer bool) {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	for _, p := range lobby.players {
		p.lock.RLock()
		if p.gameFaction == faction {
			p.lock.RUnlock()
			return p, true
		}
		p.lock.RUnlock()
	}

	return nil, false
}

func (lobby *Lobby) AddPlayer(username string, socket *websocket.Conn) (*Player, error) {
	if username == "" {
		return nil, errors.New("username cannot be blank")
	}

	if lobby.isUsernameTaken(username) {
		return nil, fmt.Errorf("username '%s' already taken", username)
	}

	lobby.lock.Lock()
	defer lobby.lock.Unlock()

	player := newPlayer(Username(username), socket, lobby.log)

	lobby.log.Infof("player '%s' joined", username)
	lobby.players = append(lobby.players, player)
	go player.readMessagesUntilSocketCloses(lobby)

	return player, nil
}

func (lobby *Lobby) RemovePlayer(username Username) {
	lobby.lock.Lock()
	defer lobby.lock.Unlock()

	for i, player := range lobby.players {
		if player.username == username {
			lobby.players = slices.Delete(lobby.players, i, i+1)
			return
		}
	}
}

func (lobby *Lobby) isUsernameTaken(username string) bool {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	for _, player := range lobby.players {
		if player.username == Username(username) {
			return true
		}
	}

	return false
}

func (lobby *Lobby) Close() {
	lobby.lock.Lock()
	defer lobby.lock.Unlock()

	for _, player := range lobby.players {
		player.lock.Lock()

		if err := player.socket.Close(); err != nil {
			player.lock.Unlock()
			player.log.ErrorCause(err, "failed to close socket connection")
		}

		player.lock.Unlock()
	}

	lobby.registry.RemoveLobby(lobby.name)

	lobby.log.Info("lobby closed")
}

func (lobby *Lobby) ClearMessages() {
	lobby.gameMessageQueue.Clear()
}

// Errors if not all player factions are selected.
func (lobby *Lobby) startGame() error {
	lobby.lock.Lock()
	defer lobby.lock.Unlock()

	claimedFactions := 0
	for _, player := range lobby.players {
		player.lock.RLock()
		if player.gameFaction != "" {
			claimedFactions++
		}
		player.lock.RUnlock()
	}
	if claimedFactions < len(lobby.game.Factions) {
		return errors.New("all player factions must be claimed before starting the game")
	}

	lobby.log.Info("starting game")
	lobby.gameStarted = true

	go func() {
		lobby.game.Run() // Runs until game is finished
		lobby.Close()
	}()

	return nil
}

func (lobby *Lobby) Logger() log.Logger {
	return lobby.log
}
