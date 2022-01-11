package app

import (
	"github.com/immerse-ntnu/hermannia/server/boards"
	"github.com/immerse-ntnu/hermannia/server/game"
	"github.com/immerse-ntnu/hermannia/server/interfaces"
)

// The global overview of games supported by the server.
var Games = map[string]interfaces.GameConstructor{
	"The Battle for Hermannia (5 players)": gameConstructor("hermannia_5players"),
}

// Returns a function to construct new game instances with the given boardName.
// The boardName must correspond to a .json file in ../game/boardconfig.
func gameConstructor(boardName string) interfaces.GameConstructor {
	return func(lobby interfaces.Lobby, options interface{}) (interfaces.Game, error) {
		board, err := boards.ReadBoard(boardName)
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
