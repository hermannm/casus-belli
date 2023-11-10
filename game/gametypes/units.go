package gametypes

import "encoding/json"

// A unit on the board, controlled by a player faction.
type Unit struct {
	Type    UnitType
	Faction PlayerFaction
}

type UnitType string

const (
	// A land unit that gets a +1 modifier in battle.
	UnitFootman UnitType = "footman"

	// A land unit that moves 2 regions at a time.
	UnitHorse UnitType = "horse"

	// A unit that can move into sea regions and coastal regions.
	UnitShip UnitType = "ship"

	// A land unit that instantly conquers neutral castles, and gets a +1 modifier in attacks on
	// castles.
	UnitCatapult UnitType = "catapult"
)

func (unit Unit) IsNone() bool {
	return unit.Type == ""
}

func (unit Unit) BattleModifier(isAttackOnCastle bool) (modifier Modifier, hasModifier bool) {
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
	if unit.IsNone() {
		return []byte("null"), nil
	}

	// Alias to avoid infinite loop of MarshalJSON.
	type unitAlias Unit

	return json.Marshal(unitAlias(unit))
}
