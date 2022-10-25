package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// A collection of players for a game.
type Lobby struct {
	name string
	game Game

	lock sync.RWMutex // Used to synchronize the adding/removal of players.

	players []*Player
}

// Represents a game instance. Used by lobbies to enable different types of games.
type Game interface {
	// Takes a game-specific player identifier string and returns a receiver to handle messages from the player.
	AddPlayer(gameID string) (GameMessageReceiver, error)

	// Returns the range of possible player IDs for this game.
	PlayerIDs() []string

	// Returns the name of the game.
	Name() string

	// Starts the game.
	Start()
}

// Receives game-specific messages for a player.
type GameMessageReceiver interface {
	ReceiveMessage(msgID string, msg json.RawMessage)
}

// Signature for functions that construct a game instance.
// Takes the lobby to which players can connect,
// and an untyped options parameter that can be parsed by the game instance for use in setup.
type GameConstructor func(lobby *Lobby, options any) (Game, error)

// Creates and registers a new lobby with the given ID,
// and uses the given constructor to construct its game instance.
// Returns error if lobby ID is already taken, or if game construction failed.
func New(name string, gameConstructor GameConstructor) (*Lobby, error) {
	lobby := &Lobby{name: name, lock: sync.RWMutex{}, players: make([]*Player, 0)}

	// TODO: Implement game options.
	game, err := gameConstructor(lobby, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make game for lobby: %w", err)
	}
	lobby.game = game

	err = lobbyRegistry.registerLobby(lobby)
	if err != nil {
		return nil, fmt.Errorf("failed to register lobby: %w", err)
	}

	return lobby, nil
}

// Marshals the given message to JSON and sends it to all connected players.
// Returns an error if it failed to marshal or send to at least one of the players.
func (lobby *Lobby) SendMessageToAll(msg map[string]any) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	var err error
	for _, player := range lobby.players {
		err = player.send(msg)
	}

	return err
}

// Finds the player of the given game ID, and attempts to send the given message to them.
func (lobby *Lobby) SendMessage(playerGameID string, msg map[string]any) error {
	player, ok := lobby.getPlayer(playerGameID)
	if !ok {
		return fmt.Errorf("failed to send message to player with id %s: player not found", playerGameID)
	}

	err := player.send(msg)
	return err
}

// Returns the player in the lobby corresponding to the given player ID, or false if none is found.
func (lobby *Lobby) getPlayer(gameID string) (*Player, bool) {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	var player *Player
	for _, p := range lobby.players {
		lobby.lock.RLock()
		if p.gameID == gameID {
			player = p
			lobby.lock.RUnlock()
			break
		}
		lobby.lock.RUnlock()
	}

	return player, player.gameID != ""
}

// Adds a player with the given username and socket to the lobby.
// Returns the player if they were successfully added to the lobby, or an error.
func (lobby *Lobby) addPlayer(username string, socket *websocket.Conn) (*Player, error) {
	if !lobby.usernameAvailable(username) {
		return nil, fmt.Errorf("username %s already taken", username)
	}

	lobby.lock.Lock()
	defer lobby.lock.Unlock()

	player := &Player{
		lock:            sync.RWMutex{},
		socket:          socket,
		username:        username,
		gameID:          "",
		gameMsgReceiver: nil,
		ready:           false,
	}

	lobby.players = append(lobby.players, player)
	go player.listen(lobby)

	return player, nil
}

func (lobby *Lobby) usernameAvailable(username string) bool {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	for _, player := range lobby.players {
		player.lock.RLock()

		if player.username == username {
			player.lock.RUnlock()
			return false
		}

		player.lock.RUnlock()
	}

	return true
}

// Removes a lobby from the lobby map and closes its connections.
func (lobby *Lobby) close() error {
	lobby.lock.Lock()
	defer lobby.lock.Unlock()

	for _, player := range lobby.players {
		player.lock.Lock()

		err := player.socket.Close()
		if err != nil {
			player.lock.Unlock()
			log.Println(fmt.Errorf("failed to close socket connection to player %s: %w", player.String(), err))
		}

		player.lock.Unlock()
	}

	lobbyRegistry.removeLobby(lobby.name)

	return nil
}

// Starts the lobby's game. Errors if not all game IDs are selected, or if not all players are ready yet.
func (lobby *Lobby) startGame() error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	claimedGameIDs := 0
	readyPlayers := 0
	for _, player := range lobby.players {
		player.lock.RLock()

		if player.gameID != "" {
			claimedGameIDs++
		}

		if player.ready {
			readyPlayers++
		}

		player.lock.RUnlock()
	}

	if claimedGameIDs < len(lobby.game.PlayerIDs()) {
		return errors.New("all game IDs must be claimed before starting the game")
	}
	if readyPlayers < len(lobby.players) {
		return errors.New("all players must mark themselves as ready before starting the game")
	}

	lobby.game.Start()

	return nil
}
