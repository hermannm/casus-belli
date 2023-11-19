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
		"Emman": {Type: game.UnitFootman, Faction: "White"},

		"Lomone": {Type: game.UnitFootman, Faction: "Green"},
		"Lusía":  {Type: game.UnitFootman, Faction: "Red"},

		"Gron":  {Type: game.UnitFootman, Faction: "White"},
		"Gnade": {Type: game.UnitFootman, Faction: "Black"},

		"Firril": {Type: game.UnitFootman, Faction: "Black"},

		"Ovo":       {Type: game.UnitFootman, Faction: "Green"},
		"Mare Elle": {Type: game.UnitShip, Faction: "Green"},

		"Winde":      {Type: game.UnitFootman, Faction: "Green"},
		"Mare Gond":  {Type: game.UnitShip, Faction: "Green"},
		"Mare Ovond": {Type: game.UnitShip, Faction: "Green"},
		"Mare Unna":  {Type: game.UnitShip, Faction: "Black"},

		"Tusser": {Type: game.UnitFootman, Faction: "White"},
		"Tige":   {Type: game.UnitFootman, Faction: "Black"},

		"Tond": {Type: game.UnitFootman, Faction: "Green"},

		"Leil":   {Type: game.UnitFootman, Faction: "Red"},
		"Limbol": {Type: game.UnitFootman, Faction: "Green"},
		"Worp":   {Type: game.UnitFootman, Faction: "Yellow"},
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
