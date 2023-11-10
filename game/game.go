package game

import (
	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/orderresolving"
	"hermannm.dev/bfh-server/game/ordervalidation"
	"hermannm.dev/wrap"
)

type Game struct {
	board     gametypes.Board
	options   GameOptions
	messenger Messenger
	Factions  []gametypes.PlayerFaction
}

type Messenger interface {
	ordervalidation.Messenger
	orderresolving.Messenger
	SendWinner(winner gametypes.PlayerFaction) error
}

func New(boardID string, options GameOptions, messenger Messenger) (*Game, error) {
	board, err := boardconfig.ReadBoardFromConfigFile(boardID)
	if err != nil {
		return nil, wrap.Error(err, "failed to create board from config file")
	}

	return &Game{
		board:     board,
		options:   options,
		messenger: messenger,
		Factions:  board.AvailablePlayerFactions(),
	}, nil
}

func (game *Game) Start() {
	season := gametypes.SeasonWinter

	for {
		orders := ordervalidation.GatherAndValidateOrders(
			game.Factions, game.board, season, game.messenger,
		)

		_, winner := orderresolving.ResolveOrders(game.board, orders, season, game.messenger)
		if winner != "" {
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
