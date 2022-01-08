package game

import (
	"errors"

	"github.com/immerse-ntnu/hermannia/server/game/messages"
)

func (game *Game) AddPlayer(playerString string) (*messages.Receiver, error) {
	player := Player(playerString)

	validPlayer := false
	for key, receiver := range game.Messages {
		if key != player {
			continue
		}

		if receiver == nil {
			return nil, errors.New("player already exists")
		}

		validPlayer = true
		break
	}
	if !validPlayer {
		return nil, errors.New("invalid player tag")
	}

	receiver := messages.NewReceiver()
	game.Messages[player] = &receiver
	return &receiver, nil
}
