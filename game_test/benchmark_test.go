package game_test

import (
	"testing"

	. "hermannm.dev/bfh-server/game"
)

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		board, round := setup()
		board.Resolve(round)
	}
}

func setup() (Board, Round) {
	units := map[string]Unit{
		"Emman": {Type: Footman, Player: "white"},

		"Limone": {Type: Footman, Player: "green"},
		"Lusía":  {Type: Footman, Player: "red"},

		"Gron":  {Type: Footman, Player: "white"},
		"Gnade": {Type: Footman, Player: "black"},

		"Firril": {Type: Footman, Player: "black"},

		"Ovo":       {Type: Footman, Player: "green"},
		"Mare Elle": {Type: Ship, Player: "green"},

		"Winde":      {Type: Footman, Player: "green"},
		"Mare Gond":  {Type: Ship, Player: "green"},
		"Mare Ovond": {Type: Ship, Player: "green"},
		"Mare Unna":  {Type: Ship, Player: "black"},

		"Tusser": {Type: Footman, Player: "white"},
		"Tige":   {Type: Footman, Player: "black"},

		"Tond": {Type: Footman, Player: "green"},

		"Leil":   {Type: Footman, Player: "red"},
		"Limbol": {Type: Footman, Player: "green"},
		"Worp":   {Type: Footman, Player: "yellow"},
	}

	orders := []Order{
		// Auto-success
		{Type: Move, From: "Emman", To: "Erren"},

		// PvP battle with defender
		{Type: Move, From: "Limone", To: "Lusía"},

		// PvP battle, no defender
		{Type: Move, From: "Gron", To: "Gewel"},
		{Type: Move, From: "Gnade", To: "Gewel"},

		// PvE battle
		{Type: Move, From: "Firril", To: "Furie"},

		// PvE battle, transport not attacked
		{Type: Move, From: "Ovo", To: "Zona"},
		{Type: Transport, From: "Mare Elle"},

		// PvE battle, transport attacked
		{Type: Move, From: "Winde", To: "Fond"},
		{Type: Transport, From: "Mare Gond"},
		{Type: Transport, From: "Mare Ovond"},
		{Type: Move, From: "Mare Unna", To: "Mare Ovond"},

		// Border battle
		{Type: Move, From: "Tusser", To: "Tige"},
		{Type: Move, From: "Tige", To: "Tusser"},

		// Danger zone, dependent move
		{Type: Move, From: "Tond", To: "Tige"},

		// Move cycle
		{Type: Move, From: "Leil", To: "Limbol"},
		{Type: Move, From: "Limbol", To: "Worp"},
		{Type: Move, From: "Worp", To: "Leil"},
	}

	board := mockBoard()
	placeUnits(board, units)

	attachUnits(orders, units)
	round := Round{FirstOrders: orders}

	return board, round
}
