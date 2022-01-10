package app

import (
	"github.com/immerse-ntnu/hermannia/server/game"
	"github.com/immerse-ntnu/hermannia/server/game/boardconfig"
	"github.com/immerse-ntnu/hermannia/server/interfaces"
)

var Games = map[string]interfaces.GameConstructor{
	"hermannia_5players": gameConstructor("hermannia_5players"),
}

func gameConstructor(boardName string) interfaces.GameConstructor {
	return func(lobby interfaces.Lobby, options interface{}) (interfaces.Game, error) {
		board, err := boardconfig.ReadBoard(boardName)
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
