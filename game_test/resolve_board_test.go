package game_test

import (
	"testing"

	. "github.com/hermannm/bfh-server/game"
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
		FirstOrders: []*Order{
			{Type: Move, Player: "red", From: board["Leil"], To: board["Limbol"]},
			{Type: Move, Player: "green", From: board["Limbol"], To: board["Worp"]},
			{Type: Move, Player: "yellow", From: board["Worp"], To: board["Leil"]},
		},
	}

	// Runs the resolve function, mutating the board.
	board.Resolve(&round)

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
