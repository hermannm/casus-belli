package game

import (
	"fmt"

	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/game/boardsetup"
	"hermannm.dev/bfh-server/game/messages"
	"hermannm.dev/bfh-server/lobby"
)

type Game struct {
	name      string
	board     board.Board
	rounds    []board.Round
	options   GameOptions
	messenger messages.Messenger
}

type GameOptions struct {
	ThroneExpansion bool // Whether the game has the "Raven, Sword and Throne" expansion enabled.
}

// Constructs a game instance. Initializes player slots for each region home tag on the given board.
func New(boardName string, options GameOptions, msgSender messages.Sender) (*Game, error) {
	brd, err := boardsetup.ReadBoard(boardName)
	if err != nil {
		return nil, err
	}

	return &Game{
		board:     brd,
		rounds:    make([]board.Round, 0),
		options:   options,
		messenger: messages.NewMessenger(msgSender),
	}, nil
}

func DefaultOptions() GameOptions {
	return GameOptions{
		ThroneExpansion: true,
	}
}

// Dynamically finds the possible player IDs for the game
// by going through the board and finding all the different Home values.
func (game Game) PlayerIDs() []string {
	return playerIDsFromBoard(game.board)
}

// Returns the name of the board played in this game.
func (game Game) Name() string {
	return game.board.Name
}

// Creates a new message receiver for the given player tag, and adds it to the game.
// Returns error if tag is invalid or already taken.
func (game Game) AddPlayer(playerID string) (lobby.GameMessageReceiver, error) {
	regionNames := make([]string, 0)
	for _, region := range game.board.Regions {
		regionNames = append(regionNames, region.Name)
	}

	receiver, err := game.messenger.AddReceiver(playerID, regionNames)
	if err != nil {
		return nil, fmt.Errorf("failed to add player: %w", err)
	}
	return receiver, nil
}

func playerIDsFromBoard(brd board.Board) []string {
	ids := make([]string, 0)

OuterLoop:
	for _, region := range brd.Regions {
		potentialID := region.HomePlayer
		if potentialID == "" {
			continue
		}

		for _, id := range ids {
			if potentialID == id {
				continue OuterLoop
			}
		}

		ids = append(ids, potentialID)
	}

	return ids
}
