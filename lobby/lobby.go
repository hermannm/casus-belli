package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Global list of game lobbies.
var lobbies = make(map[string]Lobby)

// A collection of players for a game.
type Lobby struct {
	id   string
	game Game

	lock *sync.RWMutex   // Used to synchronize the adding/removal of players.
	wg   *sync.WaitGroup // Used to wait for the lobby to fill up with players.

	// Maps player IDs (unique to the lobby) to their socket connections for sending and receiving.
	players map[string]Player
}

// A player connected to a game lobby.
type Player struct {
	id     string
	socket *websocket.Conn

	lock *sync.RWMutex // Used to synchronize reading and setting the Active field.
}

// Represents a game instance. Used by lobbies to enable different types of games.
type Game interface {
	// Takes a player identifier string (unique to this game instance, format depends on the game),
	// and returns a receiver to handle messages from the player, or an error if adding the player failed.
	AddPlayer(playerID string) (MessageReceiver, error)

	// Returns the range of possible player IDs for this game.
	PlayerIDs() []string

	// Starts the game.
	Start()
}

type MessageReceiver interface {
	ReceiveMessage(msgType string, msg json.RawMessage)
}

// Signature for functions that construct a game instance.
// Takes the lobby to which players can connect,
// and an untyped options parameter that can be parsed by the game instance for use in setup.
type GameConstructor func(lobby Lobby, options any) (Game, error)

// Creates and registers a new lobby with the given ID,
// and uses the given constructor to construct its game instance.
// Returns error if lobby ID is already taken, or if game construction failed.
func New(id string, gameConstructor GameConstructor) (Lobby, error) {
	if id == "" {
		return Lobby{}, errors.New("lobby name cannot be blank")
	}

	lobby := Lobby{id: id, lock: new(sync.RWMutex), wg: new(sync.WaitGroup), players: make(map[string]Player)}

	game, err := gameConstructor(lobby, nil)
	if err != nil {
		return Lobby{}, fmt.Errorf("failed to make game for lobby: %w", err)
	}

	lobby.game = game

	if err := registerLobby(lobby); err != nil {
		return Lobby{}, fmt.Errorf("failed to register lobby: %w", err)
	}

	return lobby, nil
}

func (lobby Lobby) activePlayers() []string {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	activePlayers := make([]string, 0)
	for playerID := range lobby.players {
		activePlayers = append(activePlayers, playerID)
	}

	return activePlayers
}

// Marshals the given message to JSON and sends it to all connected players.
// Returns an error if it failed to marshal or send to at least one of the players.
func (lobby Lobby) SendMessageToAll(msg map[string]any) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	var err error
	for _, player := range lobby.players {
		err = player.send(msg)
	}

	return err
}

func (lobby Lobby) SendMessage(to string, msg map[string]any) error {
	player, ok := lobby.getPlayer(to)
	if !ok {
		return fmt.Errorf("failed to send message to player with id %s: player not found", to)
	}

	err := player.send(msg)
	return err
}

// Marshals the given message to JSON and sends it over the player's socket connection.
// Returns an error if the player is inactive, or if the marshaling/sending failed.
func (player Player) send(msg any) error {
	player.lock.RLock()
	defer player.lock.RUnlock()

	err := player.socket.WriteJSON(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to player with id %s: %w", player.id, err)
	}
	return nil
}

// Registers a lobby in the global list of lobbies.
// Returns error if lobby with same ID already exists.
func registerLobby(lobby Lobby) error {
	if _, ok := lobbies[lobby.id]; ok {
		return errors.New("lobby with ID \"" + lobby.id + "\" already exists")
	}

	lobbies[lobby.id] = lobby
	return nil
}

// Returns the player in the lobby corresponding to the given player ID,
// or false if none is found.
func (lobby Lobby) getPlayer(playerID string) (Player, bool) {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()
	player, ok := lobby.players[playerID]
	return player, ok
}

// Sets the player in the lobby corresponding to the given player ID.
// Returns an error if no matching player is found.
func (lobby Lobby) addPlayer(player Player) error {
	lobby.lock.Lock()
	defer lobby.lock.Unlock()

	if _, taken := lobby.players[player.id]; taken {
		return fmt.Errorf("player ID %s already taken", player.id)
	}

	playerIDs := lobby.game.PlayerIDs()
	validPlayerID := false
	for _, valid := range playerIDs {
		if player.id == valid {
			validPlayerID = true
		}
	}
	if !validPlayerID {
		return errors.New("invalid player ID")
	}

	receiver, err := lobby.game.AddPlayer(player.id)
	if err != nil {
		return fmt.Errorf("failed to add player to game: %w", err)
	}

	lobby.players[player.id] = player
	go player.listen(receiver)

	return nil
}

// Returns a map of player IDs to whether they are taken (true if taken).
func (lobby Lobby) availablePlayerIDs() map[string]bool {
	available := make(map[string]bool)

	taken := lobby.activePlayers()
	total := lobby.game.PlayerIDs()

OuterLoop:
	for _, playerID := range total {
		for _, takenID := range taken {
			if playerID == takenID {
				available[playerID] = false
				continue OuterLoop
			}
		}
		available[playerID] = true
	}

	return available
}

// Removes a lobby from the lobby map and closes its connections.
func (lobby Lobby) close() error {
	for id, player := range lobby.players {
		player.lock.Lock()
		defer player.lock.Unlock()

		if err := player.socket.Close(); err != nil {
			log.Println(fmt.Errorf("failed to close socket connection to player with id %s: %w", player.id, err))
		}

		delete(lobby.players, id)
	}
	delete(lobbies, lobby.id)

	return nil
}

// If there is only 1 lobby on the server, returns that,
// otherwise returns lobby corresponding to lobby parameter in request.
// Returns error on absent lobby parameter or lobby not found.
func findLobby(req *http.Request) (Lobby, error) {
	if len(lobbies) == 1 {
		for _, lobby := range lobbies {
			return lobby, nil
		}
	}

	params, ok := checkParams(req, "lobby")
	if !ok {
		return Lobby{}, errors.New("lacking lobby query parameter")
	}

	lobbyID := params.Get("lobby")
	lobby, ok := lobbies[lobbyID]
	if !ok {
		return Lobby{}, errors.New("no lobby found with provided id")
	}

	return lobby, nil
}
