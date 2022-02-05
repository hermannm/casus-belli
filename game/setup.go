package game

import (
	"errors"

	"hermannm.dev/bfh-server/lobby"
	"hermannm.dev/bfh-server/messages"
)

// Constructs a game instance. Initializes player slots for each area home tag on the given board.
func New(board Board, lob *lobby.Lobby, options GameOptions) lobby.Game {
	receivers := make(map[Player]*messages.Receiver)
	for _, area := range board {
		if area.Home == Uncontrolled {
			continue
		}

		if _, ok := receivers[area.Home]; !ok {
			receivers[area.Home] = nil
		}
	}

	return &Game{
		Board:    board,
		Rounds:   make([]*Round, 0),
		Lobby:    lob,
		Messages: receivers,
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

// Creates a new message receiver for the given player tag, and adds it to the game.
// Returns error if tag is invalid or already taken.
func (game *Game) AddPlayer(playerID string) (lobby.Receiver, error) {
	player := Player(playerID)

	receiver, ok := game.Messages[player]
	if !ok {
		return nil, errors.New("invalid player tag")
	}
	if receiver != nil {
		return nil, errors.New("player already exists")
	}

	newReceiver := messages.NewReceiver()
	game.Messages[player] = &newReceiver
	return receiver, nil
}
