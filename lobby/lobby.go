package lobby

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
)

// A collection of players for a game.
type Lobby struct {
	name    string
	game    *game.Game
	players []*Player // Must hold lock to access safely.
	lock    sync.RWMutex
}

func New(lobbyName string, boardID string, gameOptions game.GameOptions) (*Lobby, error) {
	lobby := &Lobby{name: lobbyName, lock: sync.RWMutex{}}

	game, err := game.New(boardID, gameOptions, lobby)
	if err != nil {
		return nil, wrap.Error(err, "failed to create game")
	}
	lobby.game = game
	lobby.players = make([]*Player, 0, len(game.PlayerIDs))

	return lobby, nil
}

func (lobby *Lobby) getPlayer(gameID string) (player *Player, foundPlayer bool) {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	for _, p := range lobby.players {
		p.lock.RLock()
		if p.gameID == gameID {
			player = p
			foundPlayer = true
			p.lock.RUnlock()
			break
		}
		p.lock.RUnlock()
	}

	return player, foundPlayer
}

func (lobby *Lobby) AddPlayer(username string, socket *websocket.Conn) (*Player, error) {
	if lobby.isUsernameTaken(username) {
		return nil, fmt.Errorf("username '%s' already taken", username)
	}

	lobby.lock.Lock()
	defer lobby.lock.Unlock()

	player, err := newPlayer(username, socket)
	if err != nil {
		return nil, err
	}

	lobby.players = append(lobby.players, player)

	go player.readMessagesUntilSocketCloses(lobby)

	return player, nil
}

func (lobby *Lobby) isUsernameTaken(username string) bool {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	for _, player := range lobby.players {
		if player.username == username {
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
			log.Errorf(err, "failed to close socket connection to player %s", player.String())
		}

		player.lock.Unlock()
	}

	return nil
}

// Errors if not all game IDs are selected, or if not all players are ready yet.
func (lobby *Lobby) startGame() error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	claimedGameIDs := 0
	readyPlayers := 0
	for _, player := range lobby.players {
		player.lock.RLock()

		if player.gameID != "" {
			claimedGameIDs++
		}

		if player.readyToStartGame {
			readyPlayers++
		}

		player.lock.RUnlock()
	}

	if claimedGameIDs < len(lobby.game.PlayerIDs) {
		return errors.New("all game IDs must be claimed before starting the game")
	}
	if readyPlayers < len(lobby.players) {
		return errors.New("all players must mark themselves as ready before starting the game")
	}

	lobby.game.Start()

	return nil
}
