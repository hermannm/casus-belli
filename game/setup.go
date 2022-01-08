package game

import (
	"errors"

	"github.com/immerse-ntnu/hermannia/server/game/messages"
)

func New(board Board, lobby Lobby, options GameOptions) Game {
	messages := make(map[Player]*messages.Receiver)
	for _, area := range board {
		if area.Home == Uncontrolled {
			continue
		}

		if _, ok := messages[area.Home]; !ok {
			messages[area.Home] = nil
		}
	}

	return Game{
		Board:    board,
		Rounds:   make([]*Round, 0),
		Lobby:    lobby,
		Messages: messages,
		Options:  options,
	}
}

// Creates a new message receiver for the given player tag.
// Adds the receiver to the game, and returns it.
// Returns error if tag is invalid or already taken.
func (game *Game) AddPlayer(playerTag string) (*messages.Receiver, error) {
	player := Player(playerTag)

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
