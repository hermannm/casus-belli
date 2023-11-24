package game_test

import (
	"testing"

	"hermannm.dev/bfh-server/game"
)

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		game, orders := benchmarkSetup(b)
		b.StartTimer()
		game.ResolveNonWinterOrders(orders)
	}
}

func benchmarkSetup(b *testing.B) (*game.Game, []game.Order) {
	units := unitMap{
		"Emman": {Type: game.UnitFootman, Faction: white},

		"Furie": {Type: game.UnitHorse, Faction: black},

		"Gron":  {Type: game.UnitFootman, Faction: white},
		"Gewel": {Type: game.UnitHorse, Faction: black},

		"Lomone": {Type: game.UnitFootman, Faction: green},
		"Lusía":  {Type: game.UnitFootman, Faction: red},
		"Brodo":  {Type: game.UnitFootman, Faction: red},

		"Tusser": {Type: game.UnitFootman, Faction: white},
		"Tige":   {Type: game.UnitHorse, Faction: black},

		"Tond": {Type: game.UnitFootman, Faction: green},

		"Ovo":       {Type: game.UnitFootman, Faction: green},
		"Mare Elle": {Type: game.UnitShip, Faction: green},

		"Winde":      {Type: game.UnitFootman, Faction: green},
		"Mare Gond":  {Type: game.UnitShip, Faction: green},
		"Mare Ovond": {Type: game.UnitShip, Faction: green},
		"Mare Unna":  {Type: game.UnitShip, Faction: black},

		"Leil":   {Type: game.UnitFootman, Faction: red},
		"Limbol": {Type: game.UnitFootman, Faction: green},
		"Worp":   {Type: game.UnitFootman, Faction: yellow},
	}

	orders := []game.Order{
		// Auto-success
		{Type: game.OrderMove, Origin: "Emman", Destination: "Erren"},

		// Singleplayer battle
		{Type: game.OrderMove, Origin: "Furie", Destination: "Firril"},

		// Multiplayer battle, no defender
		{Type: game.OrderMove, Origin: "Gron", Destination: "Gnade"},
		{Type: game.OrderMove, Origin: "Gewel", Destination: "Gnade"},

		// Multiplayer battle with supported defender
		{Type: game.OrderMove, Origin: "Lomone", Destination: "Lusía"},
		{Type: game.OrderSupport, Origin: "Brodo", Destination: "Lusía"},

		// Border battle
		{Type: game.OrderMove, Origin: "Tusser", Destination: "Tige"},
		{Type: game.OrderMove, Origin: "Tige", Destination: "Tusser"},

		// Danger zone, dependent move
		{Type: game.OrderMove, Origin: "Tond", Destination: "Tige"},

		// Transport
		{Type: game.OrderMove, Origin: "Ovo", Destination: "Zona"},
		{Type: game.OrderTransport, Origin: "Mare Elle"},

		// Transport attacked
		{Type: game.OrderMove, Origin: "Winde", Destination: "Fond"},
		{Type: game.OrderTransport, Origin: "Mare Gond"},
		{Type: game.OrderTransport, Origin: "Mare Ovond"},
		{Type: game.OrderMove, Origin: "Mare Unna", Destination: "Mare Ovond"},

		// Move cycle
		{Type: game.OrderMove, Origin: "Leil", Destination: "Limbol"},
		{Type: game.OrderMove, Origin: "Limbol", Destination: "Worp"},
		{Type: game.OrderMove, Origin: "Worp", Destination: "Leil"},
	}

	game, _ := newMockGame(b, units, nil, orders, game.SeasonSpring)
	return game, orders
}
