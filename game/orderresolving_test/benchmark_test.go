package orderresolving_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/orderresolving"
	"hermannm.dev/bfh-server/game/testutils"
)

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		board, round := setup()
		orderresolving.ResolveOrders(board, round, testutils.MockMessenger{})
	}
}

func setup() (gametypes.Board, orderresolving.Round) {
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
		{Type: gametypes.OrderMove, From: "Emman", To: "Erren"},

		// PvP battle with defender
		{Type: gametypes.OrderMove, From: "Lomone", To: "Lusía"},

		// PvP battle, no defender
		{Type: gametypes.OrderMove, From: "Gron", To: "Gewel"},
		{Type: gametypes.OrderMove, From: "Gnade", To: "Gewel"},

		// PvE battle
		{Type: gametypes.OrderMove, From: "Firril", To: "Furie"},

		// PvE battle, transport not attacked
		{Type: gametypes.OrderMove, From: "Ovo", To: "Zona"},
		{Type: gametypes.OrderTransport, From: "Mare Elle"},

		// PvE battle, transport attacked
		{Type: gametypes.OrderMove, From: "Winde", To: "Fond"},
		{Type: gametypes.OrderTransport, From: "Mare Gond"},
		{Type: gametypes.OrderTransport, From: "Mare Ovond"},
		{Type: gametypes.OrderMove, From: "Mare Unna", To: "Mare Ovond"},

		// Border battle
		{Type: gametypes.OrderMove, From: "Tusser", To: "Tige"},
		{Type: gametypes.OrderMove, From: "Tige", To: "Tusser"},

		// Danger zone, dependent move
		{Type: gametypes.OrderMove, From: "Tond", To: "Tige"},

		// Move cycle
		{Type: gametypes.OrderMove, From: "Leil", To: "Limbol"},
		{Type: gametypes.OrderMove, From: "Limbol", To: "Worp"},
		{Type: gametypes.OrderMove, From: "Worp", To: "Leil"},
	}

	board := testutils.NewMockBoard()
	testutils.PlaceUnits(units, board)
	testutils.PlaceOrders(orders, board)

	round := orderresolving.Round{FirstOrders: orders}

	return board, round
}
