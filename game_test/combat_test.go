package tests

import (
	"immerse-ntnu/hermannia/server/game"
	"testing"
)

func TestAttack(test *testing.T) {
	unit := game.Unit{
		Type:  game.Footman,
		Color: game.Yellow,
	}

	area1 := game.BoardArea{
		Name:    "area1",
		Control: game.Yellow,
		Unit:    &unit,
		Forest:  false,
		Castle:  false,
		Sea:     false,
	}

	area2 := game.BoardArea{
		Name:    "area2",
		Control: game.Uncontrolled,
		Unit:    nil,
		Forest:  true,
		Castle:  true,
		Sea:     false,
	}

	area1.Neighbors["area2"] = &game.Neighbor{
		Area:        &area2,
		AcrossWater: true,
	}

	area2.Neighbors["area1"] = &game.Neighbor{
		Area:        &area1,
		AcrossWater: true,
	}

	order := game.Order{
		From: &area1,
		To:   &area2,
	}

	mods := game.AttackModifiers(order, false)
	result := 0
	for _, mod := range mods {
		result += mod.Value
	}
	expected := -2

	if result != expected {
		test.Errorf("Incorrect attack modifier, got %d, want %d", result, expected)
	}
}
