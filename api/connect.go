package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Lobby struct {
	ID          string
	PlayerCount int

	// Maps player IDs (unique to the lobby) to their socket connections for sending and receiving.
	Connections map[string]*websocket.Conn

	// Mutex for handling multiple requests trying to change the same lobby.
	Mut *sync.Mutex
}

// Sets up a new lobby with routes for joining it, one route for each given player ID.
func CreateLobby(id string, minPlayers int, playerIDs []string) {
	lobby := Lobby{
		ID:          id,
		PlayerCount: len(playerIDs),
		Connections: make(map[string]*websocket.Conn, len(playerIDs)),
	}

	for _, playerID := range playerIDs {
		http.HandleFunc("/"+id+"/join/"+playerID, addPlayer(&lobby, playerID))
	}
}

// Returns a handler for routes to add a player to the given lobby with the given player ID.
func addPlayer(lobby *Lobby, playerID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lobby.Mut.Lock()
		if _, available := lobby.Connections[playerID]; !available {
			http.Error(w, "Player ID already taken.", http.StatusConflict)
			lobby.Mut.Unlock()
			return
		}

		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Accepts all origins for now, in order to enable clients from other networks.
			CheckOrigin: func(_ *http.Request) bool { return true },
		}

		socket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unable to establish socket connection.", http.StatusInternalServerError)
		}

		lobby.Connections[playerID] = socket
		lobby.Mut.Unlock()
	}
}
