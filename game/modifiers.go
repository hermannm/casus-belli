package game

import (
	"math/rand"

	"hermannm.dev/enumnames"
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

func rollDice() int {
	return rand.Intn(6) + 1
}

func defenseModifiers(region Region) []Modifier {
	modifiers := []Modifier{
		{Type: ModifierDice, Value: rollDice()},
	}

	if unitModifier, hasModifier := region.Unit.battleModifier(false); hasModifier {
		modifiers = append(modifiers, unitModifier)
	}

	return modifiers
}

func attackModifiers(
	move Order,
	region Region,
	hasOtherAttackers bool,
	isBorderBattle bool,
	includeDefender bool,
) []Modifier {
	modifiers := []Modifier{}

	neighbor, adjacent := region.getNeighbor(move.Origin, move.ViaDangerZone)

	// Assumes danger zone checks have been made before battle,
	// and thus adds surprise modifier to attacker coming across such zones.
	if adjacent && neighbor.DangerZone != "" {
		modifiers = append(modifiers, Modifier{Type: ModifierSurprise, Value: 1})
	}

	isOnlyAttackerOnUncontrolledRegion := !region.isControlled() && !hasOtherAttackers
	isAttackOnDefendedRegion := region.isControlled() && !region.isEmpty() && includeDefender &&
		!isBorderBattle
	includeTerrainModifiers := isOnlyAttackerOnUncontrolledRegion || isAttackOnDefendedRegion

	if includeTerrainModifiers {
		if region.IsForest {
			modifiers = append(modifiers, Modifier{Type: ModifierForest, Value: -1})
		}

		if region.HasCastle {
			modifiers = append(modifiers, Modifier{Type: ModifierCastle, Value: -1})
		}

		isMovingAcrossWater := !adjacent || neighbor.IsAcrossWater
		if isMovingAcrossWater {
			modifiers = append(modifiers, Modifier{Type: ModifierWater, Value: -1})
		}
	}

	if unitModifier, hasModifier := region.Unit.battleModifier(region.HasCastle); hasModifier {
		modifiers = append(modifiers, unitModifier)
	}

	modifiers = append(modifiers, Modifier{Type: ModifierDice, Value: rollDice()})

	return modifiers
}
