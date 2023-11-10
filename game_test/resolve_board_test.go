package game_test

import (
	"testing"

	"hermannm.dev/bfh-server/game"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	units := map[game.RegionName]game.Unit{
		"Leil":   {Type: game.UnitFootman, Faction: "red"},
		"Limbol": {Type: game.UnitFootman, Faction: "green"},
		"Worp":   {Type: game.UnitFootman, Faction: "yellow"},
	}

	orders := []game.Order{
		{Type: game.OrderMove, Origin: "Leil", Destination: "Limbol"},
		{Type: game.OrderMove, Origin: "Limbol", Destination: "Worp"},
		{Type: game.OrderMove, Origin: "Worp", Destination: "Leil"},
	}

	game := newMockGame()
	placeUnits(units, game.Board)
	placeOrders(orders, game.Board)

	game.ResolveNonWinterOrders(orders)

	ExpectedControl{
		"Leil":   {ControllingFaction: "yellow", Unit: units["Worp"]},
		"Limbol": {ControllingFaction: "red", Unit: units["Leil"]},
		"Worp":   {ControllingFaction: "green", Unit: units["Limbol"]},
	}.check(game.Board, t)
}
