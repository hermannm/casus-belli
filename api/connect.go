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
}

// Sets up a new lobby with routes for joining it, one route for each given player ID.
// Returns the lobby, and a wait group to wait for the lobby to fill up.
func CreateLobby(id string, minPlayers int, playerIDs []string) (Lobby, *sync.WaitGroup) {
	lobby := Lobby{
		ID:          id,
		PlayerCount: len(playerIDs),
		Connections: make(map[string]*websocket.Conn, len(playerIDs)),
	}

	var mut sync.Mutex
	var wg sync.WaitGroup
	wg.Add(lobby.PlayerCount)

	for _, playerID := range playerIDs {
		http.HandleFunc("/"+id+"/join/"+playerID, addPlayer(&lobby, playerID, &mut, &wg))
	}

	return lobby, &wg
}

// Returns a handler for routes to add a player to the given lobby with the given player ID.
func addPlayer(lobby *Lobby, playerID string, mut *sync.Mutex, wg *sync.WaitGroup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mut.Lock()
		if _, available := lobby.Connections[playerID]; !available {
			http.Error(w, "player ID already taken", http.StatusConflict)
			mut.Unlock()
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
			http.Error(w, "unable to establish socket connection", http.StatusInternalServerError)
			mut.Unlock()
			return
		}

		lobby.Connections[playerID] = socket
		mut.Unlock()
		wg.Done()
	}
}
