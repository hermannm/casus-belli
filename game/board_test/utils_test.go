package game_test

import (
	"testing"

	. "hermannm.dev/bfh-server/game/board"
)

// Returns an empty, limited example board for testing.
func mockBoard() Board {
	board := Board{
		WinningCastleCount: 5,
	}

	areas := []Area{
		{Name: "Lusía", Castle: true},
		{Name: "Lomone", Forest: true},
		{Name: "Limbol", Forest: true},
		{Name: "Leil"},
		{Name: "Worp", Forest: true, Home: "green", Control: "green"},
		{Name: "Winde", Forest: true, Castle: true, Home: "green", Control: "green"},
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
		{Name: "Emman", Forest: true, Home: "black", Control: "black"},
		{Name: "Erren", Castle: true, Home: "black", Control: "black"},
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

	for _, area := range areas {
		area.Neighbors = make([]Neighbor, 0)
		area.IncomingMoves = make([]Order, 0)
		area.IncomingSupports = make([]Order, 0)
		board.Areas[area.Name] = area
	}

	for _, neighbor := range neighbors {
		area1 := board.Areas[neighbor.a1]
		area2 := board.Areas[neighbor.a2]

		area1.Neighbors = append(area1.Neighbors, Neighbor{
			Name:        neighbor.a2,
			AcrossWater: neighbor.river || (area1.Sea && !area2.Sea),
			Cliffs:      neighbor.cliffs,
			DangerZone:  neighbor.dangerZone,
		})
		board.Areas[neighbor.a1] = area1

		area2.Neighbors = append(area2.Neighbors, Neighbor{
			Name:        neighbor.a1,
			AcrossWater: neighbor.river || (area2.Sea && !area1.Sea),
			Cliffs:      neighbor.cliffs,
			DangerZone:  neighbor.dangerZone,
		})
		board.Areas[neighbor.a2] = area2
	}

	return board
}

// Utility function for placing units on a map.
// Takes a map of area names to units to be placed there.
func placeUnits(board Board, units map[string]Unit) {
	for areaName, unit := range units {
		area := board.Areas[areaName]
		area.Unit = unit
		area.Control = unit.Player
		board.Areas[areaName] = area
	}
}

func attachUnits(orders []Order, units map[string]Unit) {
	for i, order := range orders {
		unit, ok := units[order.From]
		if !ok {
			continue
		}

		order.Unit = unit
		order.Player = unit.Player
		orders[i] = order
	}
}

// Utility type for setting up expected outcomes of a test of board resolving.
type expectedControl map[string]struct {
	control Player
	unit    Unit
}

func checkExpectedControl(board Board, expected expectedControl, t *testing.T) {
	for name, area := range board.Areas {
		expectation, ok := expected[name]
		if !ok {
			continue
		}

		if area.Control != expectation.control {
			t.Errorf("unexpected control of %v, want %v, got %v", name, area.Control, expectation.control)
		}
		if area.Unit != expectation.unit {
			t.Errorf("unexpected unit in %v, want %v, got %v", name, area.Unit, expectation.unit)
		}
	}
}
