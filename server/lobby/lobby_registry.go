package lobby

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"hermannm.dev/condqueue"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"

	"hermannm.dev/casus-belli/server/game"
)

type LobbyRegistry struct {
	lobbies []*Lobby
	lock    sync.RWMutex
}

func NewLobbyRegistry() *LobbyRegistry {
	return &LobbyRegistry{lobbies: nil, lock: sync.RWMutex{}}
}

func (registry *LobbyRegistry) GetLobby(name string) (lobby *Lobby, lobbyFound bool) {
	registry.lock.RLock()
	defer registry.lock.RUnlock()

	for _, lobby := range registry.lobbies {
		if lobby.name == name {
			return lobby, true
		}
	}

	return nil, false
}

func (registry *LobbyRegistry) CreateLobby(
	lobbyName string,
	boardID string,
	onlyLobbyOnServer bool,
	customPlayerFactions []game.PlayerFaction,
) error {
	if lobbyName == "" {
		return errors.New("lobby name cannot be blank")
	}

	lobby := &Lobby{
		name:             lobbyName,
		players:          nil,
		game:             nil,
		gameStarted:      false,
		gameMessageQueue: condqueue.New[ReceivedMessage](),
		registry:         registry,
		lock:             sync.RWMutex{},
		log:              log.Logger{},
	}

	logger := log.Default()
	if !onlyLobbyOnServer {
		logger = logger.With("lobby", lobbyName)
	}
	lobby.log = logger

	board, boardInfo, err := game.ReadBoardFromConfigFile(boardID)
	if err != nil {
		return wrap.Error(err, "failed to read board from config file")
	}

	if len(customPlayerFactions) > 0 {
		boardInfo.PlayerFactions = customPlayerFactions
	}

	game := game.New(board, boardInfo, lobby, lobby.log, nil)
	lobby.game = game
	lobby.players = make([]*Player, 0, len(game.PlayerFactions))

	registry.lock.Lock()
	defer registry.lock.Unlock()

	for _, existingLobby := range registry.lobbies {
		if existingLobby.name == lobby.name {
			return fmt.Errorf("lobby name '%s' already taken", lobby.name)
		}
	}

	registry.lobbies = append(registry.lobbies, lobby)
	return nil
}

func (registry *LobbyRegistry) removeLobby(lobbyName string) {
	registry.lock.Lock()
	defer registry.lock.Unlock()

	for i, lobby := range registry.lobbies {
		if lobby.name == lobbyName {
			registry.lobbies = slices.Delete(registry.lobbies, i, i+1)
			return
		}
	}
}

type LobbyInfo struct {
	Name        string
	PlayerCount int
	BoardInfo   game.BoardInfo
}

func (registry *LobbyRegistry) ListLobbies() []LobbyInfo {
	registry.lock.RLock()
	defer registry.lock.RUnlock()

	lobbyList := make([]LobbyInfo, 0, len(registry.lobbies))
	for _, lobby := range registry.lobbies {
		lobby.lock.RLock()
		playerCount := len(lobby.players)
		lobby.lock.RUnlock()

		lobbyList = append(
			lobbyList,
			LobbyInfo{Name: lobby.name, PlayerCount: playerCount, BoardInfo: lobby.game.BoardInfo},
		)
	}

	return lobbyList
}
