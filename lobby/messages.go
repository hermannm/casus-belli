package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// Base for all messages.
type message map[string]any

const errorMsgID = "error"

// Message sent from server when an error occurs.
type errorMsg struct {
	Error string `json:"error"`
}

const readyMsgID = "ready"

// Message sent from client to mark themselves as ready to start the game.
type readyMsg struct {
	Ready bool `json:"ready"`
}

const startGameMsgID = "startGame"

// Message sent from lobby host to start the game once all players are ready.
type startGameMsg struct{}

// Listens for messages from the player, and forwards them to the given receiver.
// Listens continuously until the player turns inactive.
func (player Player) listen(receiver GameMessageReceiver) {
	for {
		_, receivedMsg, err := player.socket.ReadMessage()
		if err != nil {
			if err, ok := err.(*websocket.CloseError); ok {
				log.Println(fmt.Errorf("socket for player %s closed: %w", player.id, err))
				return
			}
			log.Println(fmt.Errorf("error in socket connection for player %s: %w", player.id, err))
			continue
		}

		var msgWithID map[string]json.RawMessage
		if err := json.Unmarshal(receivedMsg, &msgWithID); err != nil {
			log.Println(fmt.Errorf("failed to parse message: %w", err))
			player.sendErr("failed to parse message")
			continue
		}
		if len(msgWithID) != 1 {
			err := errors.New("invalid message format")
			log.Println(err)
			player.sendErr(err.Error())
			continue
		}

		var msgID string
		var rawMsg json.RawMessage
		for msgID, rawMsg = range msgWithID {
			break
		}

		switch msgID {
		case readyMsgID:
			continue
		case startGameMsgID:
			continue
		default:
			// If msg ID is not a lobby message ID, the message is forwarded to the game message receiver.
			go receiver.ReceiveMessage(msgID, rawMsg)
		}
	}
}

func (player Player) sendErr(errMsg string) {
	player.send(message{errorMsgID: errorMsg{Error: errMsg}})
}
