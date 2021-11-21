package utils

import (
	t "immerse-ntnu/hermannia/server/types"
)

func CombatBonus(unitType t.UnitType) int {
	switch unitType {
	case t.Footman:
		return 1
	default:
		return 0
	}
}

func DefenseModifier(area t.BoardArea) int {
	return CombatBonus(area.Unit.Type)
}

func AttackModifier(order t.Order, otherAttackers bool) int {
	attackModifier := 0

	if (order.To.Control == t.Uncontrolled && !otherAttackers) ||
		(order.To.Unit != nil && order.To.Control == order.To.Unit.Color) {
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
	}

	if neighbor, ok := order.From.Neighbors[order.To.Name]; ok {
		if neighbor.DangerZone != "" {
			attackModifier++
		}
	}

	if order.From.Unit.Type == t.Catapult && order.To.Castle {
		attackModifier++
	} else {
		attackModifier += CombatBonus(order.From.Unit.Type)
	}

	return attackModifier
}
