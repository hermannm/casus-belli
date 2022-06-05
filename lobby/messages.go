package lobby

import (
	"encoding/json"

	"hermannm.dev/bfh-server/gameserver"
)

// Lobby-specific messages from client to server.
const (
	MessageReady     = "ready"
	MessageStartGame = "startGame"
)

// Message sent from client to mark themselves as ready to start the game.
type ReadyMessage struct {
	gameserver.Message // Type: MessageReady

	Ready bool `json:"ready"`
}

// Message sent from lobby host to start the game once all players are ready.
type StartGameMessage struct {
	gameserver.Message // Type: MessageStartGame
}

// Listens for messages from the player, and forwards them to the given receiver.
// Listens continuously until the player turns inactive.
func (player *Player) Listen(receiver gameserver.MessageReceiver) {
	for {
		if !player.isActive() {
			return
		}

		_, rawMessage, err := player.Socket.ReadMessage()
		if err != nil {
			continue
		}

		var baseMessage gameserver.Message

		err = json.Unmarshal(rawMessage, &baseMessage)
		if err != nil || baseMessage.Type == "" {
			gameserver.SendError(player, "error in deserializing message")
			return
		}

		go receiver.HandleMessage(baseMessage, rawMessage)
	}
}
