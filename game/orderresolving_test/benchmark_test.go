package orderresolving_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/orderresolving"
)

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		board, orders := setup()
		orderresolving.ResolveOrders(
			board, orders, gametypes.SeasonSpring, MockMessenger{},
		)
	}
}

func setup() (gametypes.Board, []gametypes.Order) {
	units := map[string]gametypes.Unit{
		"Emman": {Type: gametypes.UnitFootman, Faction: "white"},

		"Lomone": {Type: gametypes.UnitFootman, Faction: "green"},
		"Lusía":  {Type: gametypes.UnitFootman, Faction: "red"},

		"Gron":  {Type: gametypes.UnitFootman, Faction: "white"},
		"Gnade": {Type: gametypes.UnitFootman, Faction: "black"},

		"Firril": {Type: gametypes.UnitFootman, Faction: "black"},

		"Ovo":       {Type: gametypes.UnitFootman, Faction: "green"},
		"Mare Elle": {Type: gametypes.UnitShip, Faction: "green"},

		"Winde":      {Type: gametypes.UnitFootman, Faction: "green"},
		"Mare Gond":  {Type: gametypes.UnitShip, Faction: "green"},
		"Mare Ovond": {Type: gametypes.UnitShip, Faction: "green"},
		"Mare Unna":  {Type: gametypes.UnitShip, Faction: "black"},

		"Tusser": {Type: gametypes.UnitFootman, Faction: "white"},
		"Tige":   {Type: gametypes.UnitFootman, Faction: "black"},

		"Tond": {Type: gametypes.UnitFootman, Faction: "green"},

		"Leil":   {Type: gametypes.UnitFootman, Faction: "red"},
		"Limbol": {Type: gametypes.UnitFootman, Faction: "green"},
		"Worp":   {Type: gametypes.UnitFootman, Faction: "yellow"},
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

	board := newMockBoard()
	placeUnits(units, board)
	placeOrders(orders, board)

	return board, orders
}
