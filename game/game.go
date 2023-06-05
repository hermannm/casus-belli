package game

import (
	"fmt"

	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/orderresolving"
	"hermannm.dev/bfh-server/game/ordervalidation"
)

type Game struct {
	Board     gametypes.Board
	PlayerIDs []string
	options   GameOptions
	messenger Messenger
}

type Messenger interface {
	ordervalidation.Messenger
	orderresolving.Messenger
	SendWinner(winner string) error
}

// Constructs a game instance. Initializes player slots for each region home tag on the given board.
func New(boardName string, options GameOptions, messenger Messenger) (*Game, error) {
	board, err := boardconfig.ReadBoardFromConfigFile(boardName)
	if err != nil {
		return nil, fmt.Errorf("failed to create board from config file: %w", err)
	}

	return &Game{
		Board:     board,
		PlayerIDs: board.AvailablePlayerIDs(),
		options:   options,
		messenger: messenger,
	}, nil
}

// Initializes a new round of the game.
func (game *Game) Start() {
	season := gametypes.SeasonWinter

	// Starts new rounds until there is a winner.
	for {
		orders := ordervalidation.GatherAndValidateOrders(
			game.PlayerIDs, game.Board, season, game.messenger,
		)

		_, winner, hasWinner := orderresolving.ResolveOrders(
			game.Board, orders, season, game.messenger,
		)

		if hasWinner {
			game.messenger.SendWinner(winner)
			break
		}

		season = season.Next()
	}
}

type GameOptions struct {
	ThroneExpansion bool // Whether the game has the "Raven, Sword and Throne" expansion enabled.
}

func DefaultOptions() GameOptions {
	return GameOptions{
		ThroneExpansion: true,
	}
}
