package game

import (
	"fmt"

	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/game/boardsetup"
	"hermannm.dev/bfh-server/game/messages"
	"hermannm.dev/bfh-server/lobby"
)

type Game struct {
	Board      board.Board
	Rounds     []board.Round
	Options    GameOptions
	msgHandler messages.Handler
}

type GameOptions struct {
	Thrones bool // Whether the game has the "Raven, Sword and Throne" expansion enabled.
}

// Constructs a game instance. Initializes player slots for each area home tag on the given board.
func New(boardName string, options GameOptions, msgSender messages.Sender) (*Game, error) {
	brd, err := boardsetup.ReadBoard(boardName)
	if err != nil {
		return nil, err
	}

	return &Game{
		Board:      brd,
		Rounds:     make([]board.Round, 0),
		Options:    options,
		msgHandler: messages.NewHandler(msgSender),
	}, nil
}

func DefaultOptions() GameOptions {
	return GameOptions{
		Thrones: true,
	}
}

// Dynamically finds the possible player IDs for the game
// by going through the board and finding all the different Home values.
func (game Game) PlayerIDs() []string {
	return playerIDsFromBoard(game.Board)
}

// Creates a new message receiver for the given player tag, and adds it to the game.
// Returns error if tag is invalid or already taken.
func (game Game) AddPlayer(playerID string) (lobby.MessageReceiver, error) {
	areaNames := make([]string, 0)
	for _, area := range game.Board.Areas {
		areaNames = append(areaNames, area.Name)
	}

	receiver, err := game.msgHandler.AddReceiver(playerID, areaNames)
	if err != nil {
		return nil, fmt.Errorf("failed to add player: %w", err)
	}
	return receiver, nil
}

func playerIDsFromBoard(brd board.Board) []string {
	ids := make([]string, 0)

outerLoop:
	for _, area := range brd.Areas {
		potentialID := area.HomePlayer
		if potentialID == "" {
			continue
		}

		for _, id := range ids {
			if potentialID == id {
				continue outerLoop
			}
		}

		ids = append(ids, potentialID)
	}

	return ids
}
