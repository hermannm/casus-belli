package tests

import (
	t "immerse-ntnu/hermannia/server/types"
	"immerse-ntnu/hermannia/server/utils"
	"testing"
)

func TestAttack(test *testing.T) {
	unit := t.Unit{
		Type:  t.Footman,
		Color: t.Yellow,
	}

	area1 := t.BoardArea{
		Name:    "area1",
		Control: t.Yellow,
		Unit:    &unit,
		Forest:  false,
		Castle:  false,
		Sea:     false,
	}

	area2 := t.BoardArea{
		Name:    "area2",
		Control: t.Uncontrolled,
		Unit:    nil,
		Forest:  true,
		Castle:  true,
		Sea:     false,
	}

	area1.Neighbors["area2"] = &t.Neighbor{
		Area:        &area2,
		AcrossWater: true,
	}

	area2.Neighbors["area1"] = &t.Neighbor{
		Area:        &area1,
		AcrossWater: true,
	}

	order := t.Order{
		From: &area1,
		To:   &area2,
	}

	result := utils.AttackModifier(order)
	expected := -2

	if result != expected {
		test.Errorf("Incorrect attack modifier, got %d, want %d", result, expected)
	}
}
