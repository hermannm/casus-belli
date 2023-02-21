package testutils

import (
	"testing"

	"hermannm.dev/bfh-server/game/gametypes"
)

// Returns an empty, limited example board for testing.
func NewMockBoard() gametypes.Board {
	board := gametypes.Board{
		Regions:            make(map[string]gametypes.Region),
		Name:               "Mock board",
		WinningCastleCount: 5,
	}

	regions := []gametypes.Region{
		{Name: "Lusía", Castle: true},
		{Name: "Lomone", Forest: true},
		{Name: "Limbol", Forest: true},
		{Name: "Leil"},
		{Name: "Worp", Forest: true, HomePlayer: "green", ControllingPlayer: "green"},
		{Name: "Winde", Forest: true, Castle: true, HomePlayer: "green", ControllingPlayer: "green"},
		{Name: "Ovo", Forest: true},
		{Name: "Mare Gond", Sea: true},
		{Name: "Mare Elle", Sea: true},
		{Name: "Zona"},
		{Name: "Tond"},
		{Name: "Tige"},
		{Name: "Tusser"},
		{Name: "Mare Ovond", Sea: true},
		{Name: "Furie", Castle: true},
		{Name: "Firril"},
		{Name: "Fond"},
		{Name: "Gron"},
		{Name: "Gnade"},
		{Name: "Gewel", Forest: true, Castle: true},
		{Name: "Mare Unna", Sea: true},
		{Name: "Emman", Forest: true, HomePlayer: "black", ControllingPlayer: "black"},
		{Name: "Erren", Castle: true, HomePlayer: "black", ControllingPlayer: "black"},
		{Name: "Mare Bøso", Sea: true},
	}

	// Defines a utility struct for two-way neighbor declaration, to avoid repetition.
	neighbors := []struct {
		a1         string
		a2         string
		river      bool
		cliffs     bool
		dangerZone string
	}{
		{a1: "Lusía", a2: "Lomone"},
		{a1: "Lusía", a2: "Limbol"},
		{a1: "Lusía", a2: "Leil"},
		{a1: "Lomone", a2: "Limbol"},
		{a1: "Limbol", a2: "Leil"},
		{a1: "Limbol", a2: "Worp"},
		{a1: "Leil", a2: "Worp"},
		{a1: "Leil", a2: "Winde"},
		{a1: "Leil", a2: "Ovo", river: true},
		{a1: "Worp", a2: "Winde"},
		{a1: "Worp", a2: "Mare Gond"},
		{a1: "Winde", a2: "Mare Gond"},
		{a1: "Winde", a2: "Mare Elle"},
		{a1: "Winde", a2: "Ovo", river: true},
		{a1: "Ovo", a2: "Mare Elle"},
		{a1: "Zona", a2: "Mare Elle"},
		{a1: "Zona", a2: "Mare Gond"},
		{a1: "Tond", a2: "Tige", dangerZone: "Bankene"},
		{a1: "Tond", a2: "Mare Elle"},
		{a1: "Tond", a2: "Mare Gond"},
		{a1: "Tond", a2: "Mare Ovond"},
		{a1: "Tige", a2: "Mare Elle"},
		{a1: "Tige", a2: "Mare Ovond"},
		{a1: "Tige", a2: "Tusser"},
		{a1: "Tusser", a2: "Gron", dangerZone: "Shangrila"},
		{a1: "Furie", a2: "Firril"},
		{a1: "Furie", a2: "Mare Ovond"},
		{a1: "Firril", a2: "Fond"},
		{a1: "Firril", a2: "Gron"},
		{a1: "Firril", a2: "Gnade"},
		{a1: "Firril", a2: "Mare Ovond"},
		{a1: "Fond", a2: "Mare Ovond"},
		{a1: "Fond", a2: "Mare Unna"},
		{a1: "Gron", a2: "Gnade"},
		{a1: "Gron", a2: "Gewel"},
		{a1: "Gron", a2: "Emman"},
		{a1: "Gnade", a2: "Gewel"},
		{a1: "Gewel", a2: "Mare Unna"},
		{a1: "Gewel", a2: "Emman", cliffs: true},
		{a1: "Emman", a2: "Erren", cliffs: true},
		{a1: "Emman", a2: "Mare Unna"},
		{a1: "Erren", a2: "Mare Bøso"},
		{a1: "Mare Gond", a2: "Mare Elle"},
		{a1: "Mare Gond", a2: "Mare Ovond"},
		{a1: "Mare Elle", a2: "Mare Ovond", dangerZone: "Bankene"},
		{a1: "Mare Ovond", a2: "Mare Unna"},
		{a1: "Mare Unna", a2: "Mare Bøso"},
	}

	for _, region := range regions {
		region.Neighbors = make([]gametypes.Neighbor, 0)
		region.IncomingMoves = make([]gametypes.Order, 0)
		region.IncomingSupports = make([]gametypes.Order, 0)
		board.Regions[region.Name] = region
	}

	for _, neighbor := range neighbors {
		region1 := board.Regions[neighbor.a1]
		region2 := board.Regions[neighbor.a2]

		region1.Neighbors = append(region1.Neighbors, gametypes.Neighbor{
			Name:        neighbor.a2,
			AcrossWater: neighbor.river || (region1.Sea && !region2.Sea),
			Cliffs:      neighbor.cliffs,
			DangerZone:  neighbor.dangerZone,
		})
		board.Regions[neighbor.a1] = region1

		region2.Neighbors = append(region2.Neighbors, gametypes.Neighbor{
			Name:        neighbor.a1,
			AcrossWater: neighbor.river || (region2.Sea && !region1.Sea),
			Cliffs:      neighbor.cliffs,
			DangerZone:  neighbor.dangerZone,
		})
		board.Regions[neighbor.a2] = region2
	}

	return board
}

// Utility function for placing units on the given board.
// Takes a map of region names to units to be placed there.
func PlaceUnits(units map[string]gametypes.Unit, board gametypes.Board) {
	for regionName, unit := range units {
		region := board.Regions[regionName]
		region.Unit = unit
		region.ControllingPlayer = unit.Player
		board.Regions[regionName] = region
	}
}

// Attaches units from the board to the given set of orders.
// Also sets the player field on each order to the player of the ordered unit.
func PlaceOrders(orders []gametypes.Order, board gametypes.Board) {
	for i, order := range orders {
		region, ok := board.Regions[order.From]
		if !ok {
			continue
		}

		order.Unit = region.Unit
		order.Player = region.Unit.Player
		orders[i] = order
	}
}

// Utility type for setting up expected outcomes of a test of board resolving.
type ExpectedControl map[string]struct {
	ControllingPlayer string
	Unit              gametypes.Unit
}

func (expected ExpectedControl) Check(board gametypes.Board, t *testing.T) {
	for name, region := range board.Regions {
		expectation, ok := expected[name]
		if !ok {
			continue
		}

		if region.ControllingPlayer != expectation.ControllingPlayer {
			t.Errorf(
				"unexpected control of %v, want %v, got %v",
				name,
				region.ControllingPlayer,
				expectation.ControllingPlayer,
			)
		}
		if region.Unit != expectation.Unit {
			t.Errorf("unexpected unit in %v, want %v, got %v", name, region.Unit, expectation.Unit)
		}
	}
}
