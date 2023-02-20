package gameboard_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/gameboard"
	"hermannm.dev/bfh-server/game/testutils"
)

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		board, round := setup()
		board.Resolve(round, testutils.MockMessenger{})
	}
}

func setup() (gameboard.Board, gameboard.Round) {
	units := map[string]gameboard.Unit{
		"Emman": {Type: gameboard.UnitFootman, Player: "white"},

		"Lomone": {Type: gameboard.UnitFootman, Player: "green"},
		"Lusía":  {Type: gameboard.UnitFootman, Player: "red"},

		"Gron":  {Type: gameboard.UnitFootman, Player: "white"},
		"Gnade": {Type: gameboard.UnitFootman, Player: "black"},

		"Firril": {Type: gameboard.UnitFootman, Player: "black"},

		"Ovo":       {Type: gameboard.UnitFootman, Player: "green"},
		"Mare Elle": {Type: gameboard.UnitShip, Player: "green"},

		"Winde":      {Type: gameboard.UnitFootman, Player: "green"},
		"Mare Gond":  {Type: gameboard.UnitShip, Player: "green"},
		"Mare Ovond": {Type: gameboard.UnitShip, Player: "green"},
		"Mare Unna":  {Type: gameboard.UnitShip, Player: "black"},

		"Tusser": {Type: gameboard.UnitFootman, Player: "white"},
		"Tige":   {Type: gameboard.UnitFootman, Player: "black"},

		"Tond": {Type: gameboard.UnitFootman, Player: "green"},

		"Leil":   {Type: gameboard.UnitFootman, Player: "red"},
		"Limbol": {Type: gameboard.UnitFootman, Player: "green"},
		"Worp":   {Type: gameboard.UnitFootman, Player: "yellow"},
	}

	orders := []gameboard.Order{
		// Auto-success
		{Type: gameboard.OrderMove, From: "Emman", To: "Erren"},

		// PvP battle with defender
		{Type: gameboard.OrderMove, From: "Lomone", To: "Lusía"},

		// PvP battle, no defender
		{Type: gameboard.OrderMove, From: "Gron", To: "Gewel"},
		{Type: gameboard.OrderMove, From: "Gnade", To: "Gewel"},

		// PvE battle
		{Type: gameboard.OrderMove, From: "Firril", To: "Furie"},

		// PvE battle, transport not attacked
		{Type: gameboard.OrderMove, From: "Ovo", To: "Zona"},
		{Type: gameboard.OrderTransport, From: "Mare Elle"},

		// PvE battle, transport attacked
		{Type: gameboard.OrderMove, From: "Winde", To: "Fond"},
		{Type: gameboard.OrderTransport, From: "Mare Gond"},
		{Type: gameboard.OrderTransport, From: "Mare Ovond"},
		{Type: gameboard.OrderMove, From: "Mare Unna", To: "Mare Ovond"},

		// Border battle
		{Type: gameboard.OrderMove, From: "Tusser", To: "Tige"},
		{Type: gameboard.OrderMove, From: "Tige", To: "Tusser"},

		// Danger zone, dependent move
		{Type: gameboard.OrderMove, From: "Tond", To: "Tige"},

		// Move cycle
		{Type: gameboard.OrderMove, From: "Leil", To: "Limbol"},
		{Type: gameboard.OrderMove, From: "Limbol", To: "Worp"},
		{Type: gameboard.OrderMove, From: "Worp", To: "Leil"},
	}

	board := testutils.NewMockBoard()
	testutils.PlaceUnits(units, board)
	testutils.PlaceOrders(orders, board)

	round := gameboard.Round{FirstOrders: orders}

	return board, round
}
