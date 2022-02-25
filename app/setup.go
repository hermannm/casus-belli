package app

import (
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/bfh-server/lobby"
)

// The global overview of games supported by the server.
var Games = map[string]lobby.GameConstructor{
	"The Battle for Hermannia (5 players)": gameConstructor("hermannia_5players"),
}

// Returns a function to construct new game instances with the given boardName.
// The boardName must correspond to a .json file in ../game/boardconfig.
func gameConstructor(boardName string) lobby.GameConstructor {
	return func(lob *lobby.Lobby, options interface{}) (lobby.Game, error) {
		gameOptions, ok := options.(game.GameOptions)
		if !ok {
			gameOptions = game.DefaultOptions()
		}

		return game.New(boardName, lob, gameOptions)
	}
}
