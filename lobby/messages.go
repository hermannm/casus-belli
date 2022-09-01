package lobby

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// Lobby-specific messages from client to server.
const (
	msgError     = "error"
	msgReady     = "ready"
	msgStartGame = "startGame"
)

// Base for all messages.
type message struct {
	Type string `json:"type"`
}

// Message sent from server when an error occurs.
type errorMsg struct {
	Type  string `json:"type"` // msgError
	Error string `json:"error"`
}

// Message sent from client to mark themselves as ready to start the game.
type readyMsg struct {
	Type  string `json:"type"` // msgReady
	Ready bool   `json:"ready"`
}

// Message sent from lobby host to start the game once all players are ready.
type startGameMsg struct {
	Type string `json:"type"` // msgStartGame
}

// Listens for messages from the player, and forwards them to the given receiver.
// Listens continuously until the player turns inactive.
func (player Player) listen(receiver MessageReceiver) {
	for {
		_, msg, err := player.socket.ReadMessage()
		if err != nil {
			if err, ok := err.(*websocket.CloseError); ok {
				log.Println(fmt.Errorf("socket for player %s closed: %w", player.id, err))
				return
			}
			log.Println(fmt.Errorf("error in socket connection for player %s: %w", player.id, err))
			continue
		}

		var baseMsg message

		err = json.Unmarshal(msg, &baseMsg)
		if err != nil || baseMsg.Type == "" {
			player.send(errorMsg{Type: msgError, Error: "error in deserializing message"})
			return
		}

		go receiver.ReceiveMessage(baseMsg.Type, msg)
	}
}
