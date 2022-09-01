package server

import (
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/bfh-server/lobby"
)

// The global overview of games configured for this server.
var Games = map[string]lobby.GameConstructor{
	"The Battle for Hermannia (5 players)": gameConstructor("bfh_5players"),
}

// Returns a constructor for new game instances with the given boardName.
// The boardName must correspond to a .json file in game/boardsetup/.
func gameConstructor(boardName string) lobby.GameConstructor {
	return func(lob lobby.Lobby, options any) (lobby.Game, error) {
		gameOptions, ok := options.(game.GameOptions)
		if !ok {
			gameOptions = game.DefaultOptions()
		}

		return game.New(boardName, gameOptions, lob)
	}
}
