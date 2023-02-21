package game

import (
	"fmt"

	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/messages"
	"hermannm.dev/bfh-server/game/orderresolving"
	"hermannm.dev/bfh-server/lobby"
)

type Game struct {
	board     gametypes.Board
	options   GameOptions
	messenger messages.Messenger
}

// Constructs a game instance. Initializes player slots for each region home tag on the given board.
func New(boardName string, options GameOptions, msgSender messages.Sender) (*Game, error) {
	board, err := boardconfig.ReadBoardFromConfigFile(boardName)
	if err != nil {
		return nil, err
	}

	return &Game{board: board, options: options, messenger: messages.NewMessenger(msgSender)}, nil
}

// Initializes a new round of the game.
func (game *Game) Start() {
	season := gametypes.SeasonWinter

	// Starts new rounds until there is a winner.
	for {
		orders := game.gatherAndValidateOrderSets(season)

		_, winner, hasWinner := orderresolving.ResolveOrders(
			game.board, orders, season, game.messenger,
		)

		if hasWinner {
			game.messenger.SendWinner(winner)
			break
		}

		season = season.Next()
	}
}

// Dynamically finds the possible player IDs for the game
// by going through the board and finding all the different Home values.
func (game Game) PlayerIDs() []string {
	ids := make([]string, 0)

OuterLoop:
	for _, region := range game.board.Regions {
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

type GameOptions struct {
	ThroneExpansion bool // Whether the game has the "Raven, Sword and Throne" expansion enabled.
}

func DefaultOptions() GameOptions {
	return GameOptions{
		ThroneExpansion: true,
	}
}
