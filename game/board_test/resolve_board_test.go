package board_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/game/testutils"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	units := map[string]board.Unit{
		"Leil":   {Type: board.UnitFootman, Player: "red"},
		"Limbol": {Type: board.UnitFootman, Player: "green"},
		"Worp":   {Type: board.UnitFootman, Player: "yellow"},
	}

	orders := []board.Order{
		{Type: board.OrderMove, From: "Leil", To: "Limbol"},
		{Type: board.OrderMove, From: "Limbol", To: "Worp"},
		{Type: board.OrderMove, From: "Worp", To: "Leil"},
	}

	brd := testutils.NewMockBoard()
	testutils.PlaceUnits(units, brd)
	testutils.PlaceOrders(orders, brd)

	round := board.Round{FirstOrders: orders}

	// Runs the resolve function, mutating the board.
	brd.Resolve(round, testutils.MockMessenger{})

	// Expected: the units have switched places in a circle.
	testutils.ExpectedControl{
		"Leil":   {ControllingPlayer: "yellow", Unit: units["Worp"]},
		"Limbol": {ControllingPlayer: "red", Unit: units["Leil"]},
		"Worp":   {ControllingPlayer: "green", Unit: units["Limbol"]},
	}.Check(brd, t)
}
