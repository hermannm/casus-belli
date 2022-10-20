package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// Continuously listens for messages from the player's socket until it is closed.
func (player *Player) listen(lobby *Lobby) {
	for {
		socketClosed, err := player.receiveMessage(lobby)
		if socketClosed {
			if err != nil {
				log.Println(err)
			}
			break
		}
		if err != nil {
			log.Println(fmt.Errorf("message error for player %s: %w", player.String(), err))
			player.sendErr(err.Error())
		}
	}
}

// Reads a message from the player's socket, and handles it appropriately.
// Returns socketClosed=true if the socket closed, and an error if message handling failed.
func (player *Player) receiveMessage(lobby *Lobby) (socketClosed bool, err error) {
	player.lock.RLock()
	defer player.lock.RUnlock()

	_, receivedMsg, err := player.socket.ReadMessage()
	if err != nil {
		if err, ok := err.(*websocket.CloseError); ok {
			return true, fmt.Errorf("socket closed: %w", err)
		}
		return false, fmt.Errorf("socket connection error: %w", err)
	}

	var msgWithID map[string]json.RawMessage
	if err := json.Unmarshal(receivedMsg, &msgWithID); err != nil {
		return false, fmt.Errorf("failed to parse message: %w", err)
	}
	if len(msgWithID) != 1 {
		return false, errors.New("invalid message format")
	}

	var msgID string
	var rawMsg json.RawMessage
	for msgID, rawMsg = range msgWithID {
		break
	}

	isLobbyMsg, err := player.receiveLobbyMessage(lobby, msgID, rawMsg)
	if err != nil {
		return false, fmt.Errorf("error in handling lobby message: %w", err)
	}

	// If msg ID is not a lobby message ID, the message is forwarded to the player's game message receiver.
	if !isLobbyMsg && player.gameMsgReceiver != nil {
		go player.gameMsgReceiver.ReceiveMessage(msgID, rawMsg)
	}

	return false, nil
}

// Receives a lobby-specific message, and handles it according to its ID.
func (player *Player) receiveLobbyMessage(
	lobby *Lobby, msgID string, rawMsg json.RawMessage,
) (isLobbyMsg bool, err error) {
	switch msgID {
	case selectGameIDMsgID:
		var msg selectGameIDMsg
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			return true, fmt.Errorf("failed to unmarshal %s message: %w", msgID, err)
		}

		if err := player.selectGameID(msg.GameID, lobby); err != nil {
			return true, fmt.Errorf("failed to select game ID: %w", err)
		}

		if err := lobby.sendPlayerStatusMsg(player); err != nil {
			return true, fmt.Errorf("failed to update other players about game ID selection: %w", err)
		}

		return true, nil
	case readyMsgID:
		var msg readyMsg
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			return true, fmt.Errorf("failed to unmarshal %s message: %w", msgID, err)
		}

		if err := player.setReady(msg.Ready); err != nil {
			return true, fmt.Errorf("failed to set ready status: %w", err)
		}

		if err := lobby.sendPlayerStatusMsg(player); err != nil {
			return true, fmt.Errorf("failed to update other players about ready status: %w", err)
		}

		return true, nil
	case startGameMsgID:
		if err := lobby.startGame(); err != nil {
			return true, fmt.Errorf("failed to start game: %w", err)
		}

		return true, nil
	default:
		return false, nil
	}
}
