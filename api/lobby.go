package api

import (
	"errors"
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
	Socket   *websocket.Conn
	Receiver chan []byte
	Active   bool

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
		conn.Receiver <- message
	}
}

func (conn *Connection) Receive() chan []byte {
	return conn.Receiver
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
