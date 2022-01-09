package interfaces

// Represents a game instance. Used by game servers to enable different types of games.
type Game interface {
	// Takes a player identifier string (unique to this game instance, format depends on the game),
	// and returns a receiver to handle messages from the player,
	// or an error if adding the player failed.
	AddPlayer(playerID string) (Receiver, error)
}

// Handles incoming messages from a client.
type Receiver interface {
	// Takes an unprocessed message in byte format from the client,
	// and processes it according to the game's implementation.
	// Called whenever a message is received from the client.
	HandleMessage(message []byte)
}

// Represents a lobby of players connected to a game instance.
type Lobby interface {
	// Takes a player identifier string (unique to the lobby),
	// and returns that player's connection instance,
	// or ok=false if it was not found.
	GetPlayer(playerID string) (conn Connection, ok bool)

	// Takes an untyped message, and sends it to all players connected to the lobby.
	// Returns a map of player identifiers to potential errors
	// (whole map should be nil in case of no errors).
	SendToAll(message interface{}) map[string]error

	// Closes the connections to the lobby and informs them that the lobby is closed.
	// Returns an error if closing failed.
	Close() error
}

// Handles the sending of messages to clients connected to a lobby.
type Connection interface {
	// Takes an untyped message, and sends it in some specified format to the client.
	// Returns an error if sending or formatting failed.
	Send(message interface{}) error
}

// Signature for functions that construct a game instance.
// Takes the list of player identifier strings for use by the game instance,
// the lobby to which players can connect,
// and an untyped options parameter that can be parsed by the game instance for use in setup.
type GameConstructor func(players []string, lobby Lobby, options interface{}) (Game, error)
