package game

import (
	"testing"
)

func TestResolveConflictFreeMoveCycle(t *testing.T) {
	redUnit := Unit{Type: Footman, Player: "red"}
	greenUnit := Unit{Type: Footman, Player: "green"}
	yellowUnit := Unit{Type: Footman, Player: "yellow"}

	redOrder := Order{Type: Move, Player: "red"}
	greenOrder := Order{Type: Move, Player: "green"}
	yellowOrder := Order{Type: Move, Player: "yellow"}

	board := Board{
		"Leil": &BoardArea{
			Name:          "Leil",
			Control:       "red",
			Unit:          redUnit,
			Order:         &redOrder,
			IncomingMoves: []*Order{&yellowOrder},
		},
		"Limbol": &BoardArea{
			Name:          "Limbol",
			Control:       "green",
			Unit:          greenUnit,
			Order:         &greenOrder,
			IncomingMoves: []*Order{&redOrder},
		},
		"Worp": &BoardArea{
			Name:          "Worp",
			Control:       "yellow",
			Unit:          yellowUnit,
			Order:         &yellowOrder,
			IncomingMoves: []*Order{&greenOrder},
		},
	}

	board["Leil"].Neighbors = []Neighbor{
		{
			Area: board["Limbol"],
		},
		{
			Area: board["Worp"],
		},
	}
	board["Limbol"].Neighbors = []Neighbor{
		{
			Area: board["Leil"],
		},
		{
			Area: board["Worp"],
		},
	}
	board["Worp"].Neighbors = []Neighbor{
		{
			Area: board["Limbol"],
		},
		{
			Area: board["Leil"],
		},
	}

	redOrder.From = board["Leil"]
	redOrder.To = board["Limbol"]
	greenOrder.From = board["Limbol"]
	greenOrder.To = board["Worp"]
	yellowOrder.From = board["Worp"]
	yellowOrder.To = board["Leil"]

	board.resolveMoveCycles()

	expected := map[string]struct {
		control Player
		unit    Unit
	}{
		"Leil": {
			control: "yellow",
			unit:    yellowUnit,
		},
		"Limbol": {
			control: "red",
			unit:    redUnit,
		},
		"Worp": {
			control: "green",
			unit:    greenUnit,
		},
	}

	for name, area := range board {
		if area.Control != expected[name].control {
			t.Errorf("unexpected control of %v, want %v, got %v", name, area.Control, expected[name].control)
		}
		if area.Unit != expected[name].unit {
			t.Errorf("unexpected unit in %v, want %v, got %v", name, area.Unit, expected[name].unit)
		}
	}
}
