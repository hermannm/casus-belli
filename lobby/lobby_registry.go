package lobby

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"hermannm.dev/bfh-server/game"
	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
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
) error {
	if lobbyName == "" {
		return errors.New("lobby name cannot be blank")
	}

	lobby := &Lobby{name: lobbyName, registry: registry, lock: sync.RWMutex{}}

	logger := log.Default()
	if !onlyLobbyOnServer {
		logger = logger.With(slog.String("lobby", lobbyName))
	}
	lobby.log = logger

	board, boardInfo, err := boardconfig.ReadBoardFromConfigFile(boardID)
	if err != nil {
		return wrap.Error(err, "failed to read board from config file")
	}

	game := game.New(board, boardInfo, lobby, lobby.log)
	lobby.game = game
	lobby.players = make([]*Player, 0, len(game.Factions))

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

func (registry *LobbyRegistry) RemoveLobby(lobbyName string) {
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
	Name      string
	BoardInfo game.BoardInfo
}

func (registry *LobbyRegistry) ListLobbies() []LobbyInfo {
	registry.lock.RLock()
	defer registry.lock.RUnlock()

	lobbyList := make([]LobbyInfo, 0, len(registry.lobbies))
	for _, lobby := range registry.lobbies {
		lobbyList = append(lobbyList, LobbyInfo{Name: lobby.name, BoardInfo: lobby.game.BoardInfo})
	}

	return lobbyList
}
