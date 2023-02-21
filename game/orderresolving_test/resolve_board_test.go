package orderresolving_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/orderresolving"
	"hermannm.dev/bfh-server/game/testutils"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	units := map[string]gametypes.Unit{
		"Leil":   {Type: gametypes.UnitFootman, Player: "red"},
		"Limbol": {Type: gametypes.UnitFootman, Player: "green"},
		"Worp":   {Type: gametypes.UnitFootman, Player: "yellow"},
	}

	orders := []gametypes.Order{
		{Type: gametypes.OrderMove, From: "Leil", To: "Limbol"},
		{Type: gametypes.OrderMove, From: "Limbol", To: "Worp"},
		{Type: gametypes.OrderMove, From: "Worp", To: "Leil"},
	}

	board := testutils.NewMockBoard()
	testutils.PlaceUnits(units, board)
	testutils.PlaceOrders(orders, board)

	round := orderresolving.Round{FirstOrders: orders}

	// Runs the resolve function, mutating the board.
	orderresolving.ResolveOrders(board, round, testutils.MockMessenger{})

	// Expected: the units have switched places in a circle.
	testutils.ExpectedControl{
		"Leil":   {ControllingPlayer: "yellow", Unit: units["Worp"]},
		"Limbol": {ControllingPlayer: "red", Unit: units["Leil"]},
		"Worp":   {ControllingPlayer: "green", Unit: units["Limbol"]},
	}.Check(board, t)
}
