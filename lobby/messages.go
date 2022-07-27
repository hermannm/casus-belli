package lobby

import (
	"encoding/json"
)

// Lobby-specific messages from client to server.
const (
	MsgError     = "error"
	MsgReady     = "ready"
	MsgStartGame = "startGame"
)

// Base for all messages.
type Message struct {
	Type string `json:"type"`
}

// Message sent from server when an error occurs.
type ErrorMessage struct {
	Type  string `json:"type"` // MsgError
	Error string `json:"error"`
}

// Message sent from client to mark themselves as ready to start the game.
type ReadyMessage struct {
	Type  string `json:"type"` // MsgReady
	Ready bool   `json:"ready"`
}

// Message sent from lobby host to start the game once all players are ready.
type StartGameMessage struct {
	Type string `json:"type"` // MsgStartGame
}

// Listens for messages from the player, and forwards them to the given receiver.
// Listens continuously until the player turns inactive.
func (player *Player) Listen(receiver MessageReceiver) {
	for {
		if !player.isActive() {
			return
		}

		_, msg, err := player.socket.ReadMessage()
		if err != nil {
			continue
		}

		var baseMsg Message

		err = json.Unmarshal(msg, &baseMsg)
		if err != nil || baseMsg.Type == "" {
			player.send(ErrorMessage{Type: MsgError, Error: "error in deserializing message"})
			return
		}

		go receiver.ReceiveMessage(baseMsg.Type, msg)
	}
}
