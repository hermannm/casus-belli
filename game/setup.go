package game

import (
	"errors"

	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/game/boardsetup"
	"hermannm.dev/bfh-server/game/messages"
	"hermannm.dev/bfh-server/lobby"
)

type Game struct {
	Board    board.Board
	Rounds   []board.Round
	Lobby    Lobby
	Messages map[string]messages.Receiver
	Options  GameOptions
}

type Lobby interface {
	Send(msg any) error
	GetPlayer(playerID string) (player interface{ Send(msg any) error }, ok bool)
}

type GameOptions struct {
	Thrones bool // Whether the game has the "Raven, Sword and Throne" expansion enabled.
}

// Constructs a game instance. Initializes player slots for each area home tag on the given board.
func New(boardName string, lob Lobby, options GameOptions) (*Game, error) {
	brd, err := boardsetup.ReadBoard(boardName)
	if err != nil {
		return nil, err
	}

	playerIDs := playerIDsFromBoard(brd)
	areaNames := make([]string, 0)
	for _, area := range brd.Areas {
		areaNames = append(areaNames, area.Name)
	}

	receivers := make(map[string]messages.Receiver)
	for _, playerID := range playerIDs {
		receivers[playerID] = messages.NewReceiver(areaNames)
	}

	game := Game{
		Board:    brd,
		Rounds:   make([]board.Round, 0),
		Lobby:    lob,
		Messages: receivers,
		Options:  options,
	}

	return &game, nil
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
	receiver, ok := game.Messages[playerID]
	if !ok {
		return nil, errors.New("invalid player tag")
	}

	return receiver, nil
}

func playerIDsFromBoard(brd board.Board) []string {
	ids := make([]string, 0)

outerLoop:
	for _, area := range brd.Areas {
		potentialID := string(area.Home)
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
