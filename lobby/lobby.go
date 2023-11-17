package lobby

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/condqueue"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
)

// A collection of players for a game.
type Lobby struct {
	name             string
	game             *game.Game
	players          []*Player // Must hold lock to access safely
	receivedMessages *condqueue.CondQueue[ReceivedMessage]
	lock             sync.RWMutex
	log              log.Logger
}

func New(lobbyName string, boardID string, onlyLobbyOnServer bool) (*Lobby, error) {
	lobby := &Lobby{name: lobbyName, lock: sync.RWMutex{}}

	logger := log.Default()
	if !onlyLobbyOnServer {
		logger = logger.With(slog.String("lobby", lobbyName))
	}
	lobby.log = logger

	board, boardInfo, err := boardconfig.ReadBoardFromConfigFile(boardID)
	if err != nil {
		return nil, wrap.Error(err, "failed to read board from config file")
	}
	game := game.New(board, boardInfo, lobby, lobby.log)

	lobby.game = game
	lobby.players = make([]*Player, 0, len(game.Factions))

	return lobby, nil
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

func (lobby *Lobby) Close() error {
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

	lobby.log.Info("lobby closed")
	return nil
}

// Errors if not all game IDs are selected, or if not all players are ready yet.
func (lobby *Lobby) startGame() error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	claimedFactions := 0
	readyPlayers := 0
	for _, player := range lobby.players {
		player.lock.RLock()

		if player.gameFaction != "" {
			claimedFactions++
		}

		if player.readyToStartGame {
			readyPlayers++
		}

		player.lock.RUnlock()
	}

	if claimedFactions < len(lobby.game.Factions) {
		return errors.New("all game IDs must be claimed before starting the game")
	}
	if readyPlayers < len(lobby.players) {
		return errors.New("all players must mark themselves as ready before starting the game")
	}

	lobby.log.Info("starting game")
	go lobby.game.Run()

	return nil
}

func (lobby *Lobby) Logger() log.Logger {
	return lobby.log
}
