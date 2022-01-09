package app

import (
	"github.com/immerse-ntnu/hermannia/server/game"
	"github.com/immerse-ntnu/hermannia/server/game/boardconfig"
	"github.com/immerse-ntnu/hermannia/server/interfaces"
)

var Games = map[string]interfaces.GameConstructor{
	"hermannia_5players": gameConstructor("hermannia", 5),
}

func gameConstructor(mapName string, playerCount int) interfaces.GameConstructor {
	return func(players []string, lobby interfaces.Lobby, options interface{}) (interfaces.Game, error) {
		board, err := boardconfig.ReadBoard(mapName, playerCount)
		if err != nil {
			return nil, err
		}

		gameOptions, ok := options.(game.GameOptions)
		if !ok {
			gameOptions = game.DefaultOptions()
		}

		newGame := game.New(board, lobby, gameOptions)

		return newGame, nil
	}
}
