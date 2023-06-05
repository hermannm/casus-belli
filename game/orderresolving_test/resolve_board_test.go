package orderresolving_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/orderresolving"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	units := map[string]gametypes.Unit{
		"Leil":   {Type: gametypes.UnitFootman, Player: "red"},
		"Limbol": {Type: gametypes.UnitFootman, Player: "green"},
		"Worp":   {Type: gametypes.UnitFootman, Player: "yellow"},
	}

	orders := []gametypes.Order{
		{Type: gametypes.OrderMove, Origin: "Leil", Destination: "Limbol"},
		{Type: gametypes.OrderMove, Origin: "Limbol", Destination: "Worp"},
		{Type: gametypes.OrderMove, Origin: "Worp", Destination: "Leil"},
	}

	board := newMockBoard()
	placeUnits(units, board)
	placeOrders(orders, board)

	orderresolving.ResolveOrders(board, orders, gametypes.SeasonSpring, MockMessenger{})

	ExpectedControl{
		"Leil":   {ControllingPlayer: "yellow", Unit: units["Worp"]},
		"Limbol": {ControllingPlayer: "red", Unit: units["Leil"]},
		"Worp":   {ControllingPlayer: "green", Unit: units["Limbol"]},
	}.check(board, t)
}
