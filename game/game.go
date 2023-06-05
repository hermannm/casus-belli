package game

import (
	"fmt"

	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/orderresolving"
	"hermannm.dev/bfh-server/game/ordervalidation"
)

type Game struct {
	board     gametypes.Board
	options   GameOptions
	messenger Messenger
	PlayerIDs []string
}

type Messenger interface {
	ordervalidation.Messenger
	orderresolving.Messenger
	SendWinner(winner string) error
}

func New(boardID string, options GameOptions, messenger Messenger) (*Game, error) {
	board, err := boardconfig.ReadBoardFromConfigFile(boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to create board from config file: %w", err)
	}

	return &Game{
		board:     board,
		PlayerIDs: board.AvailablePlayerIDs(),
		options:   options,
		messenger: messenger,
	}, nil
}

func (game *Game) Start() {
	season := gametypes.SeasonWinter

	for {
		orders := ordervalidation.GatherAndValidateOrders(
			game.PlayerIDs, game.board, season, game.messenger,
		)

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

func (game *Game) Name() string {
	return game.board.Name
}

type GameOptions struct {
	ThroneExpansion bool // Whether the game has the "Raven, Sword and Throne" expansion enabled.
}

func DefaultOptions() GameOptions {
	return GameOptions{
		ThroneExpansion: true,
	}
}
