package game

import (
	"hermannm.dev/enumnames"
)

// A unit on the board, controlled by a player faction.
type Unit struct {
	Type    UnitType
	Faction PlayerFaction
}

type UnitType uint8

const (
	// A land unit that gets a +1 modifier in battle.
	UnitFootman UnitType = iota + 1

	// A land unit that moves 2 regions at a time.
	UnitKnight

	// A unit that can move into sea regions and coastal regions.
	UnitShip

	// A land unit that instantly conquers neutral castles, and gets a +1 modifier in attacks on
	// castles.
	UnitCatapult
)

var unitNames = enumnames.NewMap(
	map[UnitType]string{
		UnitFootman:  "Footman",
		UnitKnight:   "Knight",
		UnitShip:     "Ship",
		UnitCatapult: "Catapult",
	},
)

func (unitType UnitType) String() string {
	return unitNames.GetNameOrFallback(unitType, "INVALID")
}

func (unitType UnitType) isValid() bool {
	return unitNames.ContainsKey(unitType)
}

func (unitType UnitType) battleModifier(
	isAttackOnCastle bool,
) (modifier Modifier, hasModifier bool) {
	modifierValue := 0
	if unitType == UnitFootman || (unitType == UnitCatapult && isAttackOnCastle) {
		modifierValue = 1
	}

	if modifierValue != 0 {
		return Modifier{Type: ModifierUnit, Value: modifierValue}, true
	} else {
		return Modifier{}, false
	}
}
