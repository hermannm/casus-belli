package lobby

import (
	"encoding/json"
)

// Lobby-specific messages from client to server.
const (
	MessageError     = "error"
	MessageReady     = "ready"
	MessageStartGame = "startGame"
)

// Base for all messages.
type Message struct {
	Type string `json:"type"`
}

// Message sent from server when an error occurs.
type ErrorMessage struct {
	Type  string `json:"type"` // MessageError
	Error string `json:"error"`
}

// Message sent from client to mark themselves as ready to start the game.
type ReadyMessage struct {
	Type  string `json:"type"` // MessageReady
	Ready bool   `json:"ready"`
}

// Message sent from lobby host to start the game once all players are ready.
type StartGameMessage struct {
	Type string `json:"type"` // MessageStartGame
}

// Listens for messages from the player, and forwards them to the given receiver.
// Listens continuously until the player turns inactive.
func (player *Player) Listen(msgHandler interface {
	HandleMessage(msgType string, msg []byte)
}) {
	for {
		if !player.isActive() {
			return
		}

		_, msg, err := player.Socket.ReadMessage()
		if err != nil {
			continue
		}

		var baseMsg Message

		err = json.Unmarshal(msg, &baseMsg)
		if err != nil || baseMsg.Type == "" {
			player.Send(ErrorMessage{Type: MessageError, Error: "error in deserializing message"})
			return
		}

		go msgHandler.HandleMessage(baseMsg.Type, msg)
	}
}
