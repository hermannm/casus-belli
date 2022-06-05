package gameserver

// Represents a game instance. Used by lobbies to enable different types of games.
type Game interface {
	// Takes a player identifier string (unique to this game instance, format depends on the game),
	// and returns a receiver to handle messages from the player,
	// or an error if adding the player failed.
	AddPlayer(playerID string) (MessageReceiver, error)

	// Returns the range of possible player IDs for this game.
	PlayerIDs() []string

	// Starts the game.
	Start()
}

// Signature for functions that construct a game instance.
// Takes the lobby to which players can connect,
// and an untyped options parameter that can be parsed by the game instance for use in setup.
type GameConstructor func(lobby Lobby, options any) (Game, error)

type Sendable interface {
	Send(message any) error
}

type Lobby interface {
	Sendable
	GetPlayer(playerID string) (Player, bool)
}

type Player interface {
	Sendable
}

// Handles game-specific messages from the client.
type MessageReceiver interface {
	// Takes the partly deserialized baseMessage for finding the message's type,
	// and the raw message for deserializing into the correct complete message type.
	HandleMessage(baseMessage Message, rawMessage []byte)
}

// Base struct for all messages to and from the server.
type Message struct {
	// Allows for correctly identifying incoming messages.
	Type string `json:"type"`
}

type ErrorMessage struct {
	Message
	Error string `json:"error"`
}

const MsgError = "error"

func SendError(to Sendable, error string) ErrorMessage {
	return ErrorMessage{Message{MsgError}, error}
}
