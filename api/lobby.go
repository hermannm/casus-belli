package api

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var lobbies = make(map[string]*Lobby)

type Lobby struct {
	ID  string
	Mut *sync.Mutex
	WG  *sync.WaitGroup

	// Maps player IDs (unique to the lobby) to their socket connections for sending and receiving.
	Connections map[string]*websocket.Conn
}

// Returns the current connected players in a lobby, and the max number of potential players.
func (lobby Lobby) PlayerCount() (current int, max int) {
	for _, conn := range lobby.Connections {
		if conn != nil {
			current++
		}
	}

	max = len(lobby.Connections)

	return current, max
}

// Returns a map of player IDs to whether they are taken (true if taken).
func (lobby Lobby) AvailablePlayerIDs() map[string]bool {
	available := make(map[string]bool)

	for playerID, conn := range lobby.Connections {
		if conn == nil {
			available[playerID] = true
		} else {
			available[playerID] = false
		}
	}

	return available
}

// Registers handlers for the lobby API routes.
func StartAPI(address string, open bool) {
	if open {
		http.HandleFunc("/new", createLobbyHandler)
	}
	http.HandleFunc("/join", addPlayer)
	http.ListenAndServe(address, nil)
}

// Creates a lobby with the given ID.
// Creates connection slot for each of the given player IDs,
// and adds an equal number to the lobby's wait group.
func CreateLobby(id string, playerIDs []string) (*Lobby, error) {
	if _, ok := lobbies[id]; ok {
		return nil, errors.New("lobby with ID \"" + id + "\" already exists")
	}

	lobby := Lobby{
		ID:          id,
		Connections: make(map[string]*websocket.Conn, len(playerIDs)),
	}
	for _, playerID := range playerIDs {
		lobby.Connections[playerID] = nil
	}
	lobby.WG.Add(len(lobby.Connections))

	lobbies[id] = &lobby

	return &lobby, nil
}

// Handler for creating lobbies for servers that let users create their own lobbies.
func createLobbyHandler(res http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()

	if !params.Has("id") || !params.Has("playerIDs") {
		http.Error(res, "insufficient query parameters", http.StatusBadRequest)
		return
	}

	id := params.Get("id")
	if _, ok := lobbies[id]; ok {
		http.Error(res, "lobby with ID \""+id+"\" already exists", http.StatusConflict)
		return
	}

	playerIDs := params["playerIDs"]
	if len(playerIDs) < 2 {
		http.Error(res, "at least 2 player IDs must be provided to lobby", http.StatusBadRequest)
		return
	}

	_, err := CreateLobby(id, playerIDs)
	if err != nil {
		http.Error(res, "error creating lobby", http.StatusInternalServerError)
		return
	}

	res.Write([]byte("lobby created"))
}

// Removes a lobby from the lobby map and closes its connections.
func CloseLobby(id string) error {
	lobby, ok := lobbies[id]
	if !ok {
		return errors.New("no lobby with ID \"" + id + "\" exists")
	}

	for _, conn := range lobby.Connections {
		conn.Close()
	}
	delete(lobbies, id)

	return nil
}

// Handler for adding a player to a lobby.
func addPlayer(res http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()

	if !params.Has("lobby") || !params.Has("player") {
		http.Error(res, "insufficient query parameters", http.StatusBadRequest)
		return
	}

	lobbyID := params.Get("lobby")
	lobby, ok := lobbies[lobbyID]
	if !ok {
		http.Error(res, "no lobby with ID "+lobbyID+" exists", http.StatusBadRequest)
	}

	playerID := params.Get("player")
	lobby.Mut.Lock()

	conn, ok := lobby.Connections[playerID]
	if !ok {
		http.Error(res, "invalid player ID", http.StatusBadRequest)
		lobby.Mut.Unlock()
		return
	}
	if conn != nil {
		http.Error(res, "player ID already taken", http.StatusConflict)
		lobby.Mut.Unlock()
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Accepts all origins for now, in order to enable clients from other networks.
		CheckOrigin: func(*http.Request) bool { return true },
	}

	socket, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println(err)
		http.Error(res, "unable to establish socket connection", http.StatusInternalServerError)
		lobby.Mut.Unlock()
		return
	}

	lobby.Connections[playerID] = socket

	res.Write([]byte("joined lobby"))

	lobby.Mut.Unlock()
	lobby.WG.Done()
}
