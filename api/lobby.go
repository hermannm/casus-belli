package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

var lobbies = make(map[string]*Lobby)

type Lobby struct {
	ID string

	Mut *sync.Mutex
	WG  *sync.WaitGroup

	// Maps player IDs (unique to the lobby) to their socket connections for sending and receiving.
	Connections map[string]Connection
}

type Connection struct {
	Socket  *websocket.Conn
	Receive chan []byte
	Active  bool

	Mut *sync.Mutex
}

func (lobby Lobby) GetConn(playerID string) (conn Connection, ok bool) {
	lobby.Mut.Lock()
	defer lobby.Mut.Unlock()
	conn, ok = lobby.Connections[playerID]
	return conn, ok
}

func (lobby Lobby) setConn(playerID string, conn Connection) error {
	lobby.Mut.Lock()
	defer lobby.Mut.Unlock()

	if _, ok := lobby.Connections[playerID]; !ok {
		return errors.New("invalid player ID")
	}

	lobby.Connections[playerID] = conn
	return nil
}

func (conn *Connection) isActive() bool {
	conn.Mut.Lock()
	defer conn.Mut.Unlock()
	return conn.Active
}

func (conn *Connection) setActive(active bool) {
	conn.Mut.Lock()
	defer conn.Mut.Unlock()
	conn.Active = active
}

func (conn *Connection) Send(message interface{}) error {
	if !conn.isActive() {
		return errors.New("cannot send to inactive connection")
	}

	err := conn.Socket.WriteJSON(message)
	return err
}

func (conn *Connection) Listen() {
	for {
		if !conn.isActive() {
			return
		}

		_, message, err := conn.Socket.ReadMessage()
		if err != nil {
			continue
		}

		conn.setActive(true)
		conn.Receive <- message
	}
}

// Returns the current connected players in a lobby, and the max number of potential players.
func (lobby Lobby) PlayerCount() (current int, max int) {
	for _, conn := range lobby.Connections {
		if conn.isActive() {
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
		if conn.isActive() {
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
	http.HandleFunc("/info", getLobby)
	http.HandleFunc("/all", getLobbies)
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
		Connections: make(map[string]Connection, len(playerIDs)),
	}
	for _, playerID := range playerIDs {
		lobby.Connections[playerID] = Connection{}
	}
	lobby.WG.Add(len(lobby.Connections))

	lobbies[id] = &lobby

	return &lobby, nil
}

// Handler for creating lobbies for servers that let users create their own lobbies.
func createLobbyHandler(res http.ResponseWriter, req *http.Request) {
	params, ok := checkParams(res, req, "id", "playerIDs")
	if !ok {
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

	for playerID, conn := range lobby.Connections {
		conn.Socket.Close()
		conn.setActive(false)
		lobby.setConn(playerID, Connection{})
	}
	delete(lobbies, id)

	return nil
}

// Handler for adding a player to a lobby.
func addPlayer(res http.ResponseWriter, req *http.Request) {
	params, ok := checkParams(res, req, "lobby", "player")
	if !ok {
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
	if conn.isActive() {
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

	conn = Connection{
		Socket: socket,
		Active: true,
	}
	lobby.Connections[playerID] = conn
	go conn.Listen()

	res.Write([]byte("joined lobby"))

	lobby.Mut.Unlock()
	lobby.WG.Done()
}

// Utility type for responding to requests for lobby info.
type lobbyInfo struct {
	ID                 string          `json:"id"`
	AvailablePlayerIDs map[string]bool `json:"availablePlayerIDs"`
}

// Handler for returning information about a given lobby.
func getLobby(res http.ResponseWriter, req *http.Request) {
	params, ok := checkParams(res, req, "lobby")
	if !ok {
		return
	}

	lobbyID := params.Get("lobby")
	lobby, ok := lobbies[lobbyID]
	if !ok {
		http.Error(res, "no lobby with id \""+lobbyID+"\"", http.StatusBadRequest)
		return
	}

	info, err := json.Marshal(lobbyInfo{
		ID:                 lobby.ID,
		AvailablePlayerIDs: lobby.AvailablePlayerIDs(),
	})
	if err != nil {
		http.Error(res, "error in reading lobby \""+lobbyID+"\"", http.StatusInternalServerError)
	}

	res.Write(info)
}

// Handler for returning information about all available lobbies.
func getLobbies(res http.ResponseWriter, req *http.Request) {
	lobbyInfoList := make([]lobbyInfo, 0, len(lobbies))

	for _, lobby := range lobbies {
		lobbyInfoList = append(lobbyInfoList, lobbyInfo{
			ID:                 lobby.ID,
			AvailablePlayerIDs: lobby.AvailablePlayerIDs(),
		})
	}

	info, err := json.Marshal(lobbyInfoList)
	if err != nil {
		http.Error(res, "error in reading lobby fetching lobby list", http.StatusInternalServerError)
	}

	res.Write(info)
}

// Checks the given request for the existence of the provided parameter keys.
// If all exist, returns the parameters, otherwise returns ok = false.
func checkParams(res http.ResponseWriter, req *http.Request, keys ...string) (
	params url.Values, ok bool,
) {
	params = req.URL.Query()

	for _, key := range keys {
		if !params.Has(key) {
			http.Error(res, "insufficient query parameters", http.StatusBadRequest)
			return nil, false
		}
	}

	return params, true
}
