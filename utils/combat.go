package utils

import (
	"immerse-ntnu/hermannia/server/types"
)

func CombatBonus(unit *types.Unit) int {
	switch unit.Type {
	case "footman":
		return 1
	default:
		return 0
	}
}

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

	if order.From.Unit.Type == types.Catapult && order.To.Castle {
		attackModifier++
	} else {
		attackModifier += CombatBonus(order.From.Unit)
	}

	return attackModifier
}
