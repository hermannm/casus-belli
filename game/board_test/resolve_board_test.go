package board_test

import (
	"testing"

	. "hermannm.dev/bfh-server/game/board"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	units := map[string]Unit{
		"Leil":   {Type: UnitFootman, Player: "red"},
		"Limbol": {Type: UnitFootman, Player: "green"},
		"Worp":   {Type: UnitFootman, Player: "yellow"},
	}

	orders := []Order{
		{Type: OrderMove, From: "Leil", To: "Limbol"},
		{Type: OrderMove, From: "Limbol", To: "Worp"},
		{Type: OrderMove, From: "Worp", To: "Leil"},
	}

	board := mockBoard()
	placeUnits(board, units)

	attachUnits(orders, units)
	round := Round{FirstOrders: orders}

	// Runs the resolve function, mutating the board.
	board.Resolve(round, mockMessageHandler{})

	// Expected: the units have switched places in a circle.
	expected := expectedControl{
		"Leil":   {controllingPlayer: "yellow", unit: units["Worp"]},
		"Limbol": {controllingPlayer: "red", unit: units["Leil"]},
		"Worp":   {controllingPlayer: "green", unit: units["Limbol"]},
	}

	checkExpectedControl(board, expected, t)
}
