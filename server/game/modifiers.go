package game

import (
	"hermannm.dev/enumnames"
	"hermannm.dev/opt"
)

// Part of a player's result in a battle.
type Modifier struct {
	Type  ModifierType
	Value int

	// Blank, unless Type is ModifierSupport.
	SupportingFaction PlayerFaction `json:",omitempty"`
}

type ModifierType uint8

// Valid values for a result modifier's type.
const (
	// Bonus from a random dice roll.
	ModifierDice ModifierType = iota + 1

	// Bonus for the type of unit.
	ModifierUnit

	// Penalty for attacking a neutral or defended forested region.
	ModifierForest

	// Penalty for attacking a neutral or defended castle region.
	ModifierCastle

	// Penalty for attacking across a river, from the sea, or across a transport.
	ModifierWater

	// Bonus for attacking across a danger zone and surviving.
	ModifierSurprise

	// Bonus from supporting player in a battle.
	ModifierSupport
)

var modifierNames = enumnames.NewMap(map[ModifierType]string{
	ModifierDice:     "Dice",
	ModifierUnit:     "Unit",
	ModifierForest:   "Forest",
	ModifierCastle:   "Castle",
	ModifierWater:    "Water",
	ModifierSurprise: "Surprise",
	ModifierSupport:  "Support",
})

func (modifierType ModifierType) String() string {
	return modifierNames.GetNameOrFallback(modifierType, "INVALID")
}

func (game *Game) newDefenderResult(unit Unit) Result {
	var modifiers []Modifier
	total := 0

	if unitModifier, hasModifier := unit.Type.battleModifier(false); hasModifier {
		modifiers = append(modifiers, unitModifier)
		total += unitModifier.Value
	}

	return Result{DefenderFaction: unit.Faction, Parts: modifiers, Total: total}
}

func (game *Game) newAttackerResult(
	move Order,
	region *Region,
	singleplayerBattle bool,
	borderBattle bool,
) Result {
	modifiers := []Modifier{}

	neighbor, adjacent := region.getNeighbor(move.Origin, move.ViaDangerZone)

	// Assumes danger zone checks have been made before battle,
	// and thus adds surprise modifier to attacker coming across such zones.
	if adjacent && neighbor.DangerZone != "" {
		modifiers = append(modifiers, Modifier{Type: ModifierSurprise, Value: 1})
	}

	isOnlyAttackerOnUncontrolledRegion := !region.controlled() && singleplayerBattle
	isAttackOnDefendedRegion := region.controlled() && !region.empty() && !borderBattle
	includeTerrainModifiers := isOnlyAttackerOnUncontrolledRegion || isAttackOnDefendedRegion

	if includeTerrainModifiers {
		if region.Forest {
			modifiers = append(modifiers, Modifier{Type: ModifierForest, Value: -1})
		}

		if region.Castle {
			modifiers = append(modifiers, Modifier{Type: ModifierCastle, Value: -1})
		}

		isMovingAcrossWater := !adjacent || neighbor.AcrossWater
		if isMovingAcrossWater {
			modifiers = append(modifiers, Modifier{Type: ModifierWater, Value: -1})
		}
	}

	if unitModifier, hasModifier := move.UnitType.battleModifier(region.Castle); hasModifier {
		modifiers = append(modifiers, unitModifier)
	}

	total := 0
	for _, modifier := range modifiers {
		total += modifier.Value
	}

	return Result{Order: opt.Value(move), Parts: modifiers, Total: total}
}
