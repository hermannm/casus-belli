package lobby

import (
	"encoding/json"
)

// Handles game-specific messages from the client.
type Receiver interface {
	// Takes the partly deserialized baseMessage for finding the message's type,
	// and the raw message for deserializing into the correct complete message type.
	HandleMessage(baseMessage Message, rawMessage []byte)
}

// Base struct for all messages to and from the server.
type Message struct {
	// Allows for correctly identifying incoming messages.
	Type string `json:"type"`
}

// Type for error messages sent from server to client.
const MessageError = "error"

type ErrorMessage struct {
	Message // Type: MessageError

	Error string `json:"error"`
}

// Lobby-specific messages from client to server.
const (
	MessageReady     = "ready"
	MessageStartGame = "startGame"
)

// Message sent from client to mark themselves as ready to start the game.
type ReadyMessage struct {
	Message // Type: MessageReady

	Ready bool `json:"ready"`
}

// Message sent from lobby host to start the game once all players are ready.
type StartGameMessage struct {
	Message // Type: MessageStartGame
}

// Listens for messages from the player, and forwards them to the given receiver.
// Listens continuously until the player turns inactive.
func (player *Player) Listen(receiver Receiver) {
	for {
		if !player.isActive() {
			return
		}

		_, rawMessage, err := player.Socket.ReadMessage()
		if err != nil {
			continue
		}

		var baseMessage Message

		err = json.Unmarshal(rawMessage, &baseMessage)
		if err != nil || baseMessage.Type == "" {
			player.Send(ErrorMessage{Message{MessageError}, "error in deserializing message"})
			return
		}

		go receiver.HandleMessage(baseMessage, rawMessage)
	}
}
