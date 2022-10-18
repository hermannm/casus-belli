package lobby

import (
	"errors"
	"fmt"
	"sync"
)

var lobbyRegistry = NewLobbyRegistry()

type LobbyRegistry struct {
	lock sync.RWMutex

	lobbies map[string]*Lobby
}

func NewLobbyRegistry() *LobbyRegistry {
	return &LobbyRegistry{
		lock:    sync.RWMutex{},
		lobbies: make(map[string]*Lobby),
	}
}

// Returns the lobby of the given name from the registry, or false if it is not found.
func (registry *LobbyRegistry) getLobby(name string) (*Lobby, bool) {
	registry.lock.RLock()
	defer registry.lock.RUnlock()

	lobby, ok := registry.lobbies[name]
	return lobby, ok
}

// Attempts to add the given lobby to the lobby registry. Errors if lobby name is invalid or already taken.
func (registry *LobbyRegistry) registerLobby(lobby *Lobby) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	if lobby.name == "" {
		return errors.New("lobby name cannot be blank")
	}

	registry.lock.Lock()
	defer registry.lock.Unlock()

	uniqueName := true
	for _, existingLobby := range registry.lobbies {
		existingLobby.lock.RLock()
		if existingLobby.name == lobby.name {
			uniqueName = false
		}
		existingLobby.lock.RUnlock()
	}
	if !uniqueName {
		return fmt.Errorf("lobby name \"%s\" already taken", lobby.name)
	}

	registry.lobbies[lobby.name] = lobby

	return nil
}

// Removes the lobby of the given name from the lobby registry.
func (registry *LobbyRegistry) removeLobby(lobbyName string) {
	registry.lock.Lock()
	defer registry.lock.Unlock()

	delete(registry.lobbies, lobbyName)
}

type lobbyInfo struct {
	LobbyName string `json:"lobbyName"`
	GameName  string `json:"gameName"`
}

// Returns an info object for each lobby in the registry.
func (registry *LobbyRegistry) lobbyInfo() []lobbyInfo {
	registry.lock.RLock()
	defer registry.lock.RUnlock()

	info := make([]lobbyInfo, 0, len(registry.lobbies))

	for _, lobby := range registry.lobbies {
		lobby.lock.RLock()

		info = append(info, lobbyInfo{LobbyName: lobby.name, GameName: lobby.game.Name()})

		lobby.lock.RUnlock()
	}

	return info
}
