package tests

import (
	. "immerse-ntnu/hermannia/server/types"
	"immerse-ntnu/hermannia/server/utils"
	"testing"
)

func TestAttack(test *testing.T) {
	unit := Unit{
		Type:  Footman,
		Color: Yellow,
	}

	area1 := BoardArea{
		Name:    "area1",
		Control: Yellow,
		Unit:    &unit,
		Forest:  false,
		Castle:  false,
		Sea:     false,
	}

	area2 := BoardArea{
		Name:    "area2",
		Control: Uncontrolled,
		Unit:    nil,
		Forest:  true,
		Castle:  true,
		Sea:     false,
	}

	area1.Neighbors["area2"] = &Neighbor{
		Area:        &area2,
		AcrossWater: true,
	}

	area2.Neighbors["area1"] = &Neighbor{
		Area:        &area1,
		AcrossWater: true,
	}

	order := Order{
		From: &area1,
		To:   &area2,
	}

	result := utils.AttackModifier(order, false)
	expected := -2

	if result != expected {
		test.Errorf("Incorrect attack modifier, got %d, want %d", result, expected)
	}
}
