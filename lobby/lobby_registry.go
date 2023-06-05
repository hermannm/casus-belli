package lobby

import (
	"errors"
	"fmt"
	"sync"
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

func (registry *LobbyRegistry) RegisterLobby(lobby *Lobby) error {
	if lobby.name == "" {
		return errors.New("lobby name cannot be blank")
	}

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

	remainingLobbies := make([]*Lobby, 0, cap(registry.lobbies))
	for _, lobby := range registry.lobbies {
		if lobby.name != lobbyName {
			remainingLobbies = append(remainingLobbies, lobby)
		}
	}

	registry.lobbies = remainingLobbies
}

type LobbyInfo struct {
	LobbyName string `json:"lobbyName"`
	GameName  string `json:"gameName"`
}

func (registry *LobbyRegistry) LobbyInfo() []LobbyInfo {
	registry.lock.RLock()
	defer registry.lock.RUnlock()

	info := make([]LobbyInfo, 0, len(registry.lobbies))
	for _, lobby := range registry.lobbies {
		info = append(info, LobbyInfo{LobbyName: lobby.name, GameName: lobby.game.Name()})
	}

	return info
}
