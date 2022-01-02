package api

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

// Global list of game lobbies.
var lobbies = make(map[string]*Lobby)

// A collection of players for a game.
type Lobby struct {
	ID   string
	Game Game

	Mut *sync.Mutex     // Used to synchronize the adding/removal of players.
	WG  *sync.WaitGroup // Used to wait for the lobby to fill up with players.

	// Maps player IDs (unique to the lobby) to their socket connections for sending and receiving.
	Connections map[string]Connection
}

// A player's connection to a game lobby.
type Connection struct {
	Socket   *websocket.Conn
	Active   bool // Whether the connection is initialized/not timed out.
	Receiver Receiver

	Mut *sync.Mutex // Used to synchronize reading and setting the Active field.
}

type Receiver interface {
	HandleMessage([]byte)
}

type Game interface {
	AddPlayer(string) (Receiver, error)
}

// Returns the player connection in the lobby corresponding to the given player ID,
// or ok=false if none is found.
func (lobby Lobby) GetPlayer(playerID string) (conn Connection, ok bool) {
	lobby.Mut.Lock()
	defer lobby.Mut.Unlock()
	conn, ok = lobby.Connections[playerID]
	return conn, ok
}

// Sets the player connection in the lobby corresponding to the given player ID.
// Returns an error if no matching player is found.
func (lobby Lobby) setPlayer(playerID string, conn Connection) error {
	lobby.Mut.Lock()
	defer lobby.Mut.Unlock()

	if _, ok := lobby.Connections[playerID]; !ok {
		return errors.New("invalid player ID")
	}

	lobby.Connections[playerID] = conn
	return nil
}

// Returns the Active flag of a connection in a thread-safe manner.
func (conn *Connection) isActive() bool {
	conn.Mut.Lock()
	defer conn.Mut.Unlock()
	return conn.Active
}

// Sets the Active flag of a connection in a thread-safe manner.
func (conn *Connection) setActive(active bool) {
	conn.Mut.Lock()
	defer conn.Mut.Unlock()
	conn.Active = active
}

// Marshals the given message to JSON and sends it over the connection.
// Returns an error if the connection is inactive, or if the marshaling/sending failed.
func (conn *Connection) Send(message interface{}) error {
	if !conn.isActive() {
		return errors.New("cannot send to inactive connection")
	}

	err := conn.Socket.WriteJSON(message)
	return err
}

func (lobby *Lobby) SendToAll(message interface{}) []error {
	var errs []error

	for _, conn := range lobby.Connections {
		err := conn.Send(message)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// Listens for messages from the connection, and forwards them to the connection's receiver channel.
// Listens continuously until the connection turns inactive.
func (conn *Connection) Listen() {
	for {
		if !conn.isActive() {
			return
		}

		_, message, err := conn.Socket.ReadMessage()
		if err != nil {
			continue
		}

		go conn.Receiver.HandleMessage(message)
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
func (lobby Lobby) Close() error {
	for playerID, conn := range lobby.Connections {
		conn.Socket.Close()
		conn.setActive(false)
		lobby.setPlayer(playerID, Connection{})
	}
	delete(lobbies, lobby.ID)

	return nil
}
