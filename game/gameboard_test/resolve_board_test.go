package gameboard_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/gameboard"
	"hermannm.dev/bfh-server/game/testutils"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	units := map[string]gameboard.Unit{
		"Leil":   {Type: gameboard.UnitFootman, Player: "red"},
		"Limbol": {Type: gameboard.UnitFootman, Player: "green"},
		"Worp":   {Type: gameboard.UnitFootman, Player: "yellow"},
	}

	orders := []gameboard.Order{
		{Type: gameboard.OrderMove, From: "Leil", To: "Limbol"},
		{Type: gameboard.OrderMove, From: "Limbol", To: "Worp"},
		{Type: gameboard.OrderMove, From: "Worp", To: "Leil"},
	}

	board := testutils.NewMockBoard()
	testutils.PlaceUnits(units, board)
	testutils.PlaceOrders(orders, board)

	round := gameboard.Round{FirstOrders: orders}

	// Runs the resolve function, mutating the board.
	board.Resolve(round, testutils.MockMessenger{})

	// Expected: the units have switched places in a circle.
	testutils.ExpectedControl{
		"Leil":   {ControllingPlayer: "yellow", Unit: units["Worp"]},
		"Limbol": {ControllingPlayer: "red", Unit: units["Leil"]},
		"Worp":   {ControllingPlayer: "green", Unit: units["Limbol"]},
	}.Check(board, t)
}
