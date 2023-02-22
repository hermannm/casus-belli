package orderresolving_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/orderresolving"
	"hermannm.dev/bfh-server/game/testutils"
)

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		board, orders := setup()
		orderresolving.ResolveOrders(
			board, orders, gametypes.SeasonSpring, testutils.MockMessenger{},
		)
	}
}

func setup() (gametypes.Board, []gametypes.Order) {
	units := map[string]gametypes.Unit{
		"Emman": {Type: gametypes.UnitFootman, Player: "white"},

		"Lomone": {Type: gametypes.UnitFootman, Player: "green"},
		"Lusía":  {Type: gametypes.UnitFootman, Player: "red"},

		"Gron":  {Type: gametypes.UnitFootman, Player: "white"},
		"Gnade": {Type: gametypes.UnitFootman, Player: "black"},

		"Firril": {Type: gametypes.UnitFootman, Player: "black"},

		"Ovo":       {Type: gametypes.UnitFootman, Player: "green"},
		"Mare Elle": {Type: gametypes.UnitShip, Player: "green"},

		"Winde":      {Type: gametypes.UnitFootman, Player: "green"},
		"Mare Gond":  {Type: gametypes.UnitShip, Player: "green"},
		"Mare Ovond": {Type: gametypes.UnitShip, Player: "green"},
		"Mare Unna":  {Type: gametypes.UnitShip, Player: "black"},

		"Tusser": {Type: gametypes.UnitFootman, Player: "white"},
		"Tige":   {Type: gametypes.UnitFootman, Player: "black"},

		"Tond": {Type: gametypes.UnitFootman, Player: "green"},

		"Leil":   {Type: gametypes.UnitFootman, Player: "red"},
		"Limbol": {Type: gametypes.UnitFootman, Player: "green"},
		"Worp":   {Type: gametypes.UnitFootman, Player: "yellow"},
	}

	orders := []gametypes.Order{
		// Auto-success
		{Type: gametypes.OrderMove, Origin: "Emman", Destination: "Erren"},

		// PvP battle with defender
		{Type: gametypes.OrderMove, Origin: "Lomone", Destination: "Lusía"},

		// PvP battle, no defender
		{Type: gametypes.OrderMove, Origin: "Gron", Destination: "Gewel"},
		{Type: gametypes.OrderMove, Origin: "Gnade", Destination: "Gewel"},

		// PvE battle
		{Type: gametypes.OrderMove, Origin: "Firril", Destination: "Furie"},

		// PvE battle, transport not attacked
		{Type: gametypes.OrderMove, Origin: "Ovo", Destination: "Zona"},
		{Type: gametypes.OrderTransport, Origin: "Mare Elle"},

		// PvE battle, transport attacked
		{Type: gametypes.OrderMove, Origin: "Winde", Destination: "Fond"},
		{Type: gametypes.OrderTransport, Origin: "Mare Gond"},
		{Type: gametypes.OrderTransport, Origin: "Mare Ovond"},
		{Type: gametypes.OrderMove, Origin: "Mare Unna", Destination: "Mare Ovond"},

		// Border battle
		{Type: gametypes.OrderMove, Origin: "Tusser", Destination: "Tige"},
		{Type: gametypes.OrderMove, Origin: "Tige", Destination: "Tusser"},

		// Danger zone, dependent move
		{Type: gametypes.OrderMove, Origin: "Tond", Destination: "Tige"},

		// Move cycle
		{Type: gametypes.OrderMove, Origin: "Leil", Destination: "Limbol"},
		{Type: gametypes.OrderMove, Origin: "Limbol", Destination: "Worp"},
		{Type: gametypes.OrderMove, Origin: "Worp", Destination: "Leil"},
	}

	board := testutils.NewMockBoard()
	testutils.PlaceUnits(units, board)
	testutils.PlaceOrders(orders, board)

	return board, orders
}
