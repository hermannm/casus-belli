package gametypes

import "encoding/json"

// A player unit on the board.
type Unit struct {
	// Affects how the unit moves and its battle capabilities.
	Type UnitType `json:"unit"`

	// The player owning the unit.
	Player string `json:"player"`
}

// Type of player unit on the board (affects how it moves and its battle capabilities).
// See UnitType constants for possible values.
type UnitType string

// Valid values for a player unit's type.
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

// Checks whether the unit is initialized.
func (unit Unit) IsNone() bool {
	return unit.Type == ""
}

func (unit Unit) BattleModifier(isAttackOnCastle bool) (modifier Modifier, hasModifier bool) {
	modifierValue := 0
	switch unit.Type {
	case UnitFootman:
		modifierValue = 1
	case UnitCatapult:
		modifierValue = 1
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
