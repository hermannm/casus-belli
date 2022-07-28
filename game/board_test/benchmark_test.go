package board_test

import (
	"testing"

	. "hermannm.dev/bfh-server/game/board"
)

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		board, round := setup()
		board.Resolve(round, mockMessageHandler{})
	}
}

func setup() (Board, Round) {
	units := map[string]Unit{
		"Emman": {Type: UnitFootman, Player: "white"},

		"Lomone": {Type: UnitFootman, Player: "green"},
		"Lusía":  {Type: UnitFootman, Player: "red"},

		"Gron":  {Type: UnitFootman, Player: "white"},
		"Gnade": {Type: UnitFootman, Player: "black"},

		"Firril": {Type: UnitFootman, Player: "black"},

		"Ovo":       {Type: UnitFootman, Player: "green"},
		"Mare Elle": {Type: UnitShip, Player: "green"},

		"Winde":      {Type: UnitFootman, Player: "green"},
		"Mare Gond":  {Type: UnitShip, Player: "green"},
		"Mare Ovond": {Type: UnitShip, Player: "green"},
		"Mare Unna":  {Type: UnitShip, Player: "black"},

		"Tusser": {Type: UnitFootman, Player: "white"},
		"Tige":   {Type: UnitFootman, Player: "black"},

		"Tond": {Type: UnitFootman, Player: "green"},

		"Leil":   {Type: UnitFootman, Player: "red"},
		"Limbol": {Type: UnitFootman, Player: "green"},
		"Worp":   {Type: UnitFootman, Player: "yellow"},
	}

	orders := []Order{
		// Auto-success
		{Type: OrderMove, From: "Emman", To: "Erren"},

		// PvP battle with defender
		{Type: OrderMove, From: "Lomone", To: "Lusía"},

		// PvP battle, no defender
		{Type: OrderMove, From: "Gron", To: "Gewel"},
		{Type: OrderMove, From: "Gnade", To: "Gewel"},

		// PvE battle
		{Type: OrderMove, From: "Firril", To: "Furie"},

		// PvE battle, transport not attacked
		{Type: OrderMove, From: "Ovo", To: "Zona"},
		{Type: OrderTransport, From: "Mare Elle"},

		// PvE battle, transport attacked
		{Type: OrderMove, From: "Winde", To: "Fond"},
		{Type: OrderTransport, From: "Mare Gond"},
		{Type: OrderTransport, From: "Mare Ovond"},
		{Type: OrderMove, From: "Mare Unna", To: "Mare Ovond"},

		// Border battle
		{Type: OrderMove, From: "Tusser", To: "Tige"},
		{Type: OrderMove, From: "Tige", To: "Tusser"},

		// Danger zone, dependent move
		{Type: OrderMove, From: "Tond", To: "Tige"},

		// Move cycle
		{Type: OrderMove, From: "Leil", To: "Limbol"},
		{Type: OrderMove, From: "Limbol", To: "Worp"},
		{Type: OrderMove, From: "Worp", To: "Leil"},
	}

	board := mockBoard()
	placeUnits(board, units)

	attachUnits(orders, units)
	round := Round{FirstOrders: orders}

	return board, round
}
