package game_test

import (
	"testing"

	. "hermannm.dev/bfh-server/game"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	board := mockBoard()

	units := map[string]Unit{
		"Leil":   {Type: Footman, Player: "red"},
		"Limbol": {Type: Footman, Player: "green"},
		"Worp":   {Type: Footman, Player: "yellow"},
	}
	placeUnits(board, units)

	round := Round{
		FirstOrders: []Order{
			{Type: Move, Player: "red", From: "Leil", To: "Limbol", Unit: board["Leil"].Unit},
			{Type: Move, Player: "green", From: "Limbol", To: "Worp", Unit: board["Limbol"].Unit},
			{Type: Move, Player: "yellow", From: "Worp", To: "Leil", Unit: board["Worp"].Unit},
		},
	}

	// Runs the resolve function, mutating the board.
	board.Resolve(round)

	// Expected: the units have switched places in a circle.
	expected := expectedControl{
		"Leil": {
			control: "yellow",
			unit:    units["Worp"],
		},
		"Limbol": {
			control: "red",
			unit:    units["Leil"],
		},
		"Worp": {
			control: "green",
			unit:    units["Limbol"],
		},
	}

	checkExpectedControl(board, expected, t)
}
