package utils

import (
	"immerse/hermannia/server/types"
)

func AttackModifier(order types.Order) int {
	attackModifier := 0

	if order.To.Forest {
		attackModifier--
	}
	if order.To.Castle {
		attackModifier--
	}
	if neighbor, ok := order.From.Neighbors[order.To.Name]; ok {
		if neighbor.AcrossWater {
			attackModifier--
		}
	}

	if order.From.OccupyingUnit.Type == types.Catapult && order.To.Castle {
		attackModifier++
	} else {
		attackModifier += order.From.OccupyingUnit.CombatBonus()
	}

	return attackModifier
}
