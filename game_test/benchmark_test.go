package game_test

import (
	"testing"

	"hermannm.dev/bfh-server/game"
)

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		game, orders := setup()
		game.ResolveNonWinterOrders(orders)
	}
}

func setup() (*game.Game, []game.Order) {
	units := map[game.RegionName]game.Unit{
		"Emman": {Type: game.UnitFootman, Faction: "white"},

		"Lomone": {Type: game.UnitFootman, Faction: "green"},
		"Lusía":  {Type: game.UnitFootman, Faction: "red"},

		"Gron":  {Type: game.UnitFootman, Faction: "white"},
		"Gnade": {Type: game.UnitFootman, Faction: "black"},

		"Firril": {Type: game.UnitFootman, Faction: "black"},

		"Ovo":       {Type: game.UnitFootman, Faction: "green"},
		"Mare Elle": {Type: game.UnitShip, Faction: "green"},

		"Winde":      {Type: game.UnitFootman, Faction: "green"},
		"Mare Gond":  {Type: game.UnitShip, Faction: "green"},
		"Mare Ovond": {Type: game.UnitShip, Faction: "green"},
		"Mare Unna":  {Type: game.UnitShip, Faction: "black"},

		"Tusser": {Type: game.UnitFootman, Faction: "white"},
		"Tige":   {Type: game.UnitFootman, Faction: "black"},

		"Tond": {Type: game.UnitFootman, Faction: "green"},

		"Leil":   {Type: game.UnitFootman, Faction: "red"},
		"Limbol": {Type: game.UnitFootman, Faction: "green"},
		"Worp":   {Type: game.UnitFootman, Faction: "yellow"},
	}

	orders := []game.Order{
		// Auto-success
		{Type: game.OrderMove, Origin: "Emman", Destination: "Erren"},

		// PvP battle with defender
		{Type: game.OrderMove, Origin: "Lomone", Destination: "Lusía"},

		// PvP battle, no defender
		{Type: game.OrderMove, Origin: "Gron", Destination: "Gewel"},
		{Type: game.OrderMove, Origin: "Gnade", Destination: "Gewel"},

		// PvE battle
		{Type: game.OrderMove, Origin: "Firril", Destination: "Furie"},

		// PvE battle, transport not attacked
		{Type: game.OrderMove, Origin: "Ovo", Destination: "Zona"},
		{Type: game.OrderTransport, Origin: "Mare Elle"},

		// PvE battle, transport attacked
		{Type: game.OrderMove, Origin: "Winde", Destination: "Fond"},
		{Type: game.OrderTransport, Origin: "Mare Gond"},
		{Type: game.OrderTransport, Origin: "Mare Ovond"},
		{Type: game.OrderMove, Origin: "Mare Unna", Destination: "Mare Ovond"},

		// Border battle
		{Type: game.OrderMove, Origin: "Tusser", Destination: "Tige"},
		{Type: game.OrderMove, Origin: "Tige", Destination: "Tusser"},

		// Danger zone, dependent move
		{Type: game.OrderMove, Origin: "Tond", Destination: "Tige"},

		// Move cycle
		{Type: game.OrderMove, Origin: "Leil", Destination: "Limbol"},
		{Type: game.OrderMove, Origin: "Limbol", Destination: "Worp"},
		{Type: game.OrderMove, Origin: "Worp", Destination: "Leil"},
	}

	game := newMockGame()
	placeUnits(units, game.Board)
	placeOrders(orders, game.Board)

	return game, orders
}
