package game_test

import (
	"testing"

	"hermannm.dev/bfh-server/game"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	units := map[game.RegionName]game.Unit{
		"Leil":   {Type: game.UnitFootman, Faction: "Red"},
		"Limbol": {Type: game.UnitFootman, Faction: "Green"},
		"Worp":   {Type: game.UnitFootman, Faction: "Yellow"},
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
		"Leil":   {ControllingFaction: "Yellow", Unit: units["Worp"]},
		"Limbol": {ControllingFaction: "Red", Unit: units["Leil"]},
		"Worp":   {ControllingFaction: "Green", Unit: units["Limbol"]},
	}.check(game.Board, t)
}
