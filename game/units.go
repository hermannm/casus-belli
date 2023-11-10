package game

import (
	"encoding/json"

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
	UnitHorse

	// A unit that can move into sea regions and coastal regions.
	UnitShip

	// A land unit that instantly conquers neutral castles, and gets a +1 modifier in attacks on
	// castles.
	UnitCatapult
)

var unitNames = enumnames.NewMap(map[UnitType]string{
	UnitFootman:  "Footman",
	UnitHorse:    "Horse",
	UnitShip:     "Ship",
	UnitCatapult: "Catapult",
})

func (unitType UnitType) String() string {
	return unitNames.GetNameOrFallback(unitType, "INVALID")
}

func (unitType UnitType) isNone() bool {
	return unitType == 0
}

func (unit Unit) isNone() bool {
	return unit.Type.isNone()
}

func (unit Unit) battleModifier(isAttackOnCastle bool) (modifier Modifier, hasModifier bool) {
	modifierValue := 0
	switch unit.Type {
	case UnitFootman:
		modifierValue = 1
	case UnitCatapult:
		if isAttackOnCastle {
			modifierValue = 1
		}
	}

	if modifierValue != 0 {
		return Modifier{Type: ModifierUnit, Value: modifierValue}, true
	} else {
		return Modifier{}, false
	}
}

// Custom json.Marshaler implementation, to serialize uninitialized units to null.
func (unit Unit) MarshalJSON() ([]byte, error) {
	if unit.isNone() {
		return []byte("null"), nil
	}

	// Alias to avoid infinite loop of MarshalJSON.
	type unitAlias Unit

	return json.Marshal(unitAlias(unit))
}
