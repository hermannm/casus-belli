package board_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/game/testutils"
)

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		board, round := setup()
		board.Resolve(round, testutils.MockMessenger{})
	}
}

func setup() (board.Board, board.Round) {
	units := map[string]board.Unit{
		"Emman": {Type: board.UnitFootman, Player: "white"},

		"Lomone": {Type: board.UnitFootman, Player: "green"},
		"Lusía":  {Type: board.UnitFootman, Player: "red"},

		"Gron":  {Type: board.UnitFootman, Player: "white"},
		"Gnade": {Type: board.UnitFootman, Player: "black"},

		"Firril": {Type: board.UnitFootman, Player: "black"},

		"Ovo":       {Type: board.UnitFootman, Player: "green"},
		"Mare Elle": {Type: board.UnitShip, Player: "green"},

		"Winde":      {Type: board.UnitFootman, Player: "green"},
		"Mare Gond":  {Type: board.UnitShip, Player: "green"},
		"Mare Ovond": {Type: board.UnitShip, Player: "green"},
		"Mare Unna":  {Type: board.UnitShip, Player: "black"},

		"Tusser": {Type: board.UnitFootman, Player: "white"},
		"Tige":   {Type: board.UnitFootman, Player: "black"},

		"Tond": {Type: board.UnitFootman, Player: "green"},

		"Leil":   {Type: board.UnitFootman, Player: "red"},
		"Limbol": {Type: board.UnitFootman, Player: "green"},
		"Worp":   {Type: board.UnitFootman, Player: "yellow"},
	}

	orders := []board.Order{
		// Auto-success
		{Type: board.OrderMove, From: "Emman", To: "Erren"},

		// PvP battle with defender
		{Type: board.OrderMove, From: "Lomone", To: "Lusía"},

		// PvP battle, no defender
		{Type: board.OrderMove, From: "Gron", To: "Gewel"},
		{Type: board.OrderMove, From: "Gnade", To: "Gewel"},

		// PvE battle
		{Type: board.OrderMove, From: "Firril", To: "Furie"},

		// PvE battle, transport not attacked
		{Type: board.OrderMove, From: "Ovo", To: "Zona"},
		{Type: board.OrderTransport, From: "Mare Elle"},

		// PvE battle, transport attacked
		{Type: board.OrderMove, From: "Winde", To: "Fond"},
		{Type: board.OrderTransport, From: "Mare Gond"},
		{Type: board.OrderTransport, From: "Mare Ovond"},
		{Type: board.OrderMove, From: "Mare Unna", To: "Mare Ovond"},

		// Border battle
		{Type: board.OrderMove, From: "Tusser", To: "Tige"},
		{Type: board.OrderMove, From: "Tige", To: "Tusser"},

		// Danger zone, dependent move
		{Type: board.OrderMove, From: "Tond", To: "Tige"},

		// Move cycle
		{Type: board.OrderMove, From: "Leil", To: "Limbol"},
		{Type: board.OrderMove, From: "Limbol", To: "Worp"},
		{Type: board.OrderMove, From: "Worp", To: "Leil"},
	}

	brd := testutils.NewMockBoard()
	testutils.PlaceUnits(units, brd)
	testutils.PlaceOrders(orders, brd)

	round := board.Round{FirstOrders: orders}

	return brd, round
}
