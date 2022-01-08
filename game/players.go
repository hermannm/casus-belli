package game

import (
	"errors"

	"github.com/immerse-ntnu/hermannia/server/game/messages"
)

func (game *Game) AddPlayer(playerString string) (*messages.Receiver, error) {
	player := Player(playerString)

	receiver, ok := game.Messages[player]
	if !ok {
		return nil, errors.New("invalid player tag")
	}
	if receiver != nil {
		return nil, errors.New("player already exists")
	}

	newReceiver := messages.NewReceiver()
	game.Messages[player] = &newReceiver
	return &newReceiver, nil
}
