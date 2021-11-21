package tests

import (
	"immerse/hermannia/server/types"
	"immerse/hermannia/server/utils"
	"testing"
)

func TestAttack(t *testing.T) {
	unit := types.Unit{
		Type:  types.Footman,
		Color: types.Yellow,
	}

	area1 := types.BoardArea{
		Name:          "area1",
		Control:       types.Yellow,
		OccupyingUnit: &unit,
		Forest:        false,
		Castle:        false,
		Sea:           false,
	}

	area2 := types.BoardArea{
		Name:          "area2",
		OccupyingUnit: nil,
		Forest:        true,
		Castle:        true,
		Sea:           false,
	}

	area1.Neighbors["area2"] = &types.Neighbor{
		Area:        &area2,
		AcrossWater: true,
	}

	area2.Neighbors["area1"] = &types.Neighbor{
		Area:        &area1,
		AcrossWater: true,
	}

	order := types.Order{
		From: &area1,
		To:   &area2,
	}

	result := utils.AttackModifier(order)
	expected := -2

	if result != expected {
		t.Errorf("Incorrect attack modifier, got %d, want %d", result, expected)
	}
}
