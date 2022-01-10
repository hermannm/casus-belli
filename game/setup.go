package game

import (
	"errors"

	"github.com/immerse-ntnu/hermannia/server/game/messages"
	"github.com/immerse-ntnu/hermannia/server/interfaces"
)

func New(board Board, lobby interfaces.Lobby, options GameOptions) interfaces.Game {
	messages := make(map[Player]*messages.Receiver)
	for _, area := range board {
		if area.Home == Uncontrolled {
			continue
		}

		if _, ok := messages[area.Home]; !ok {
			messages[area.Home] = nil
		}
	}

	return &Game{
		Board:    board,
		Rounds:   make([]*Round, 0),
		Lobby:    lobby,
		Messages: messages,
		Options:  options,
	}
}

func DefaultOptions() GameOptions {
	return GameOptions{
		Thrones: true,
	}
}

// Dynamically finds the possible player IDs for the game
// by going through the board and finding all the different Home values.
func (game Game) PlayerIDs() []string {
	ids := make([]string, 0)

outer:
	for _, area := range game.Board {
		if area.Home == Uncontrolled {
			continue
		}

		potentialID := string(area.Home)

		for _, id := range ids {
			if potentialID == id {
				continue outer
			}
		}

		ids = append(ids, potentialID)
	}

	return ids
}

// Creates a new message receiver for the given player tag.
// Adds the receiver to the game, and returns it.
// Returns error if tag is invalid or already taken.
func (game *Game) AddPlayer(playerTag string) (interfaces.Receiver, error) {
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
