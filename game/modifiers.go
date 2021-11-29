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

// Returns modifiers (including dice roll) of defending unit in an area.
func DefenseModifiers(area BoardArea) []Modifier {
	mods := []Modifier{}

	mods = AppendUnitMod(mods, area.Unit.Type)

	mods = append(mods, DiceModifier())

	return mods
}

// Returns modifiers (including dice roll) of attacking unit in an area.
func AttackModifiers(order Order, otherAttackers bool, borderConflict bool) []Modifier {
	mods := []Modifier{}

	neighbor, hasNeighbor := order.From.GetNeighbor(order.To.Name, order.Via)

	// Assumes danger zone checks have been made before combat,
	// and thus adds surprise modifier to attacker coming across such zones.
	if hasNeighbor {
		if neighbor.DangerZone != "" {
			mods = append(mods, Modifier{
				Type:  SurpriseMod,
				Value: +1,
			})
		}
	}

	// Terrain modifiers should be added if:
	// - Area is uncontrolled, and this unit is the only attacker.
	// - Destination is controlled and defended, and this is not a border conflict.
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

		if hasNeighbor {
			// Attacker takes penalty for moving across river or from sea.
			if neighbor.River || (order.From.Sea && !order.To.Sea) {
				mods = append(mods, Modifier{
					Type:  WaterMod,
					Value: -1,
				})
			}
		} else {
			// If origin and destination are not neighbors, then attacker is transported,
			// and takes penalty for moving across water.
			mods = append(mods, Modifier{
				Type:  WaterMod,
				Value: -1,
			})
		}
	}

	// Catapults get a bonus only in attacks on castle areas.
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

// Returns a pseudo-random integer between 1 and 6.
func RollDice() int {
	// Uses nanoseconds since 1970 as random seed generator, to approach random outcome.
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(6) + 1
}
