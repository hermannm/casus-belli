package game

import (
	"math/rand"
	"time"
)

// Returns modifiers (including dice roll) of defending unit in an area.
func defenseModifiers(area BoardArea) []Modifier {
	mods := []Modifier{}

	mods = appendUnitMod(mods, area.Unit.Type)

	mods = append(mods, diceModifier())

	return mods
}

// Returns modifiers (including dice roll) of attacking unit in an area.
func attackModifiers(order Order, otherAttackers bool, borderConflict bool) []Modifier {
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
		(!order.To.IsEmpty() && order.To.Control == order.To.Unit.Color && !borderConflict) {

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
		mods = appendUnitMod(mods, order.From.Unit.Type)
	}

	mods = append(mods, diceModifier())

	return mods
}

// Appends unit modifier to the list if given unit type provides a modifier.
func appendUnitMod(mods []Modifier, unitType UnitType) []Modifier {
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

// Rolls dice and wraps result in a modifier.
func diceModifier() Modifier {
	return Modifier{
		Type:  DiceMod,
		Value: rollDice(),
	}
}

// Returns a pseudo-random integer between 1 and 6.
func rollDice() int {
	// Uses nanoseconds since 1970 as random seed generator, to approach random outcome.
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(6) + 1
}

// Calls support from support orders to the given area.
// Appends support modifiers to receiving players' modifier lists in the given map.
func appendSupportMods(mods map[PlayerColor][]Modifier, area BoardArea, moves []*Order) {
	for _, support := range area.IncomingSupports {
		supported := callSupport(support, area, moves)

		if _, isPlayer := mods[supported]; isPlayer {
			mods[supported] = append(mods[supported], Modifier{
				Type:        SupportMod,
				Value:       1,
				SupportFrom: support.Player,
			})
		}
	}
}

// Returns which player a given support order supports in a combat.
// If combatant matches support order's player, support is automatically given.
// If support is not given to any combatant, returns "".
//
// TODO: Implement asking player who to support if they are not involved themselves.
func callSupport(support *Order, area BoardArea, moves []*Order) PlayerColor {
	if !area.IsEmpty() && area.Unit.Color == support.Player {
		return support.Player
	}

	for _, move := range moves {
		if support.Player == move.From.Unit.Color {
			return support.Player
		}
	}

	return ""
}

// Constructs combat results from combatants' modifiers.
func combatResults(playerMods map[PlayerColor][]Modifier) (
	combat Combat,
	winner Result,
	tie bool,
) {
	for player, mods := range playerMods {
		total := sumModifiers(mods)

		result := Result{
			Total:  sumModifiers(mods),
			Parts:  mods,
			Player: player,
		}

		if total > winner.Total {
			winner = result
			tie = false
		} else if total == winner.Total {
			tie = true
		}

		combat = append(combat, result)
	}

	return combat, winner, tie
}

func sumModifiers(mods []Modifier) int {
	total := 0
	for _, mod := range mods {
		total += mod.Value
	}
	return total
}
