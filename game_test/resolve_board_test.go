package game_test

import (
	"testing"

	. "hermannm.dev/bfh-server/game"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	units := map[string]Unit{
		"Leil":   {Type: Footman, Player: "red"},
		"Limbol": {Type: Footman, Player: "green"},
		"Worp":   {Type: Footman, Player: "yellow"},
	}

	orders := []Order{
		{Type: Move, From: "Leil", To: "Limbol"},
		{Type: Move, From: "Limbol", To: "Worp"},
		{Type: Move, From: "Worp", To: "Leil"},
	}

	board := mockBoard()
	placeUnits(board, units)

	attachUnits(orders, units)
	round := Round{FirstOrders: orders}

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
