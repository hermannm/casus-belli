package lobby

import (
	"errors"
	"net/http"
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
	Players map[string]*Player
}

// A player's connection to a game lobby.
type Player struct {
	Socket   *websocket.Conn
	Active   bool // Whether the connection is initialized/not timed out.
	Receiver Receiver

	Mut *sync.Mutex // Used to synchronize reading and setting the Active field.
}

// Represents a game instance. Used by lobbies to enable different types of games.
type Game interface {
	// Takes a player identifier string (unique to this game instance, format depends on the game),
	// and returns a receiver to handle messages from the player,
	// or an error if adding the player failed.
	AddPlayer(playerID string) (Receiver, error)

	// Returns the range of possible player IDs for this game.
	PlayerIDs() []string
}

// Signature for functions that construct a game instance.
// Takes the lobby to which players can connect,
// and an untyped options parameter that can be parsed by the game instance for use in setup.
type GameConstructor func(lobby *Lobby, options interface{}) (Game, error)

// Handles incoming messages from a client.
type Receiver interface {
	// Takes an unprocessed message in byte format from the client,
	// and processes it according to the game's implementation.
	// Called whenever a message is received from the client.
	HandleMessage(message []byte)
}

// Returns the player connection in the lobby corresponding to the given player ID,
// or false if none is found.
func (lobby *Lobby) GetPlayer(playerID string) (*Player, bool) {
	lobby.Mut.Lock()
	defer lobby.Mut.Unlock()
	player, ok := lobby.Players[playerID]
	return player, ok
}

// Sets the player connection in the lobby corresponding to the given player ID.
// Returns an error if no matching player is found.
func (lobby Lobby) setPlayer(playerID string, conn Player) error {
	lobby.Mut.Lock()
	defer lobby.Mut.Unlock()

	if _, ok := lobby.Players[playerID]; !ok {
		return errors.New("invalid player ID")
	}

	lobby.Players[playerID] = &conn
	return nil
}

// Returns the Active flag of a connection in a thread-safe manner.
func (conn *Player) isActive() bool {
	conn.Mut.Lock()
	defer conn.Mut.Unlock()
	return conn.Active
}

// Sets the Active flag of a connection in a thread-safe manner.
func (conn *Player) setActive(active bool) {
	conn.Mut.Lock()
	defer conn.Mut.Unlock()
	conn.Active = active
}

// Marshals the given message to JSON and sends it over the connection.
// Returns an error if the connection is inactive, or if the marshaling/sending failed.
func (conn *Player) Send(message interface{}) error {
	if !conn.isActive() {
		return errors.New("cannot send to inactive connection")
	}

	err := conn.Socket.WriteJSON(message)
	return err
}

// Takes an untyped message, and sends it to all players connected to the lobby.
// Returns a map of player identifiers to potential errors
// (whole map should be nil in case of no errors).
func (lobby *Lobby) SendToAll(message interface{}) map[string]error {
	var errs map[string]error

	for id, conn := range lobby.Players {
		err := conn.Send(message)
		if err != nil {
			if errs == nil {
				errs = make(map[string]error)
			}

			errs[id] = err
		}
	}

	return errs
}

// Listens for messages from the connection, and forwards them to the connection's receiver channel.
// Listens continuously until the connection turns inactive.
func (conn *Player) Listen() {
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
	for _, conn := range lobby.Players {
		if conn.isActive() {
			current++
		}
	}

	max = len(lobby.Players)

	return current, max
}

// Returns a map of player IDs to whether they are taken (true if taken).
func (lobby Lobby) AvailablePlayerIDs() map[string]bool {
	available := make(map[string]bool)

	for playerID, conn := range lobby.Players {
		if conn.isActive() {
			available[playerID] = true
		} else {
			available[playerID] = false
		}
	}

	return available
}

// Creates and registers a new lobby with the given ID and constructed game instance.
// Returns error if lobby ID is already taken, or if game construction failed.
func New(id string, gameConstructor GameConstructor) (*Lobby, error) {
	if id == "" {
		return nil, errors.New("lobby name cannot be blank")
	}

	lobby := Lobby{
		ID: id,
	}

	game, err := gameConstructor(&lobby, nil)
	if err != nil {
		return nil, err
	}

	lobby.Game = game
	playerIDs := game.PlayerIDs()
	lobby.AddPlayerSlots(playerIDs)

	err = RegisterLobby(&lobby)
	if err != nil {
		return nil, err
	}

	return &lobby, nil
}

// Takes the given list of player IDs and adds connection slots for each of them in the lobby.
// Adds the length of the given IDs to the lobby's wait group, so it can be used to wait for the lobby to fill up.
func (lobby *Lobby) AddPlayerSlots(playerIDs []string) {
	lobby.Players = make(map[string]*Player, len(playerIDs))
	for _, playerID := range playerIDs {
		lobby.Players[playerID] = &Player{}
	}
	var wg sync.WaitGroup
	wg.Add(len(playerIDs))
	lobby.WG = &wg
}

// Registers a lobby in the global list of lobbies.
// Returns error if lobby with same ID already exists.
func RegisterLobby(lobby *Lobby) error {
	if _, ok := lobbies[lobby.ID]; ok {
		return errors.New("lobby with ID \"" + lobby.ID + "\" already exists")
	}

	lobbies[lobby.ID] = lobby
	return nil
}

// Removes a lobby from the lobby map and closes its connections.
func (lobby Lobby) Close() error {
	for playerID, conn := range lobby.Players {
		conn.Socket.Close()
		conn.setActive(false)
		lobby.setPlayer(playerID, Player{})
	}
	delete(lobbies, lobby.ID)

	return nil
}

// If there is only 1 lobby on the server, returns that,
// otherwise returns lobby corresponding to lobby parameter in request.
// Returns error on absent lobby parameter or lobby not found.
func findLobby(req *http.Request) (*Lobby, error) {
	if len(lobbies) == 1 {
		for _, lobby := range lobbies {
			return lobby, nil
		}
	}

	params, ok := checkParams(req, "lobby")
	if !ok {
		return nil, errors.New("lacking lobby query parameter")
	}

	lobbyID := params.Get("lobby")
	lobby, ok := lobbies[lobbyID]
	if !ok {
		return nil, errors.New("no lobby found with provided id")
	}

	return lobby, nil
}
