package game

import (
	"math/rand"
	"time"
)

func modTotal(mods []Modifier) int {
	total := 0
	for _, mod := range mods {
		total += mod.Value
	}
	return total
}

func AppendUnitMod(mods []Modifier, unitType UnitType) []Modifier {
	switch unitType {
	case Footman:
		return append(mods, Modifier{
			Type:  UnitMod,
			Value: +1,
		})
	default:
		return mods
	}
}

func DefenseModifiers(area BoardArea) []Modifier {
	mods := []Modifier{}

	mods = AppendUnitMod(mods, area.Unit.Type)

	mods = append(mods, DiceModifier())

	return mods
}

func AttackModifiers(order Order, otherAttackers bool, borderConflict bool) []Modifier {
	mods := []Modifier{}

	neighbor, hasNeighbor := order.From.GetNeighbor(order.To.Name, order.Via)

	if hasNeighbor {
		mods = append(mods, Modifier{
			Type:  SurpriseMod,
			Value: +1,
		})
	}

	if (order.To.Control == Uncontrolled && !otherAttackers) ||
		(order.To.Unit != nil && order.To.Control == order.To.Unit.Color && !borderConflict) {

		if order.To.Forest {
			mods = append(mods, Modifier{
				Type:  ForestMod,
				Value: -1,
			})
		}

		if order.To.Castle {
			mods = append(mods, Modifier{
				Type:  CastleMod,
				Value: -1,
			})
		}

		if hasNeighbor && neighbor.DangerZone != "" {
			if neighbor.River || (order.From.Sea && !order.To.Sea) {
				mods = append(mods, Modifier{
					Type:  WaterMod,
					Value: -1,
				})
			}
		} else {
			/* If destination is not in neighbors, then order is transported,
			and takes penalty for moving across water */
			mods = append(mods, Modifier{
				Type:  WaterMod,
				Value: -1,
			})
		}
	}

	if order.From.Unit.Type == Catapult && order.To.Castle {
		mods = append(mods, Modifier{
			Type:  UnitMod,
			Value: +1,
		})
	} else {
		mods = AppendUnitMod(mods, order.From.Unit.Type)
	}

	mods = append(mods, DiceModifier())

	return mods
}

func DiceModifier() Modifier {
	return Modifier{
		Type:  DiceMod,
		Value: RollDice(),
	}
}

func RollDice() int {
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(6) + 1
}
