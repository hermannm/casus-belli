package gametypes

import (
	"math/rand"
	"time"
)

// A typed number that adds to a player's result in a battle.
type Modifier struct {
	// The source of the modifier.
	Type ModifierType `json:"type"`

	// The positive or negative number that modifies the result total.
	Value int `json:"value"`

	// If modifier was from a support: the supporting player.
	SupportingPlayer string `json:"supportingPlayer"`
}

// The source of a modifier.
type ModifierType string

// Valid values for a result modifier's type.
const (
	// Bonus from a random dice roll.
	ModifierDice ModifierType = "dice"

	// Bonus for the type of unit.
	ModifierUnit ModifierType = "unit"

	// Penalty for attacking a neutral or defended forested region.
	ModifierForest ModifierType = "forest"

	// Penalty for attacking a neutral or defended castle region.
	ModifierCastle ModifierType = "castle"

	// Penalty for attacking across a river, from the sea, or across a transport.
	ModifierWater ModifierType = "water"

	// Bonus for attacking across a danger zone and surviving.
	ModifierSurprise ModifierType = "surprise"

	// Bonus from supporting player in a battle.
	ModifierSupport ModifierType = "support"
)

func RollDiceBonus() Modifier {
	// Uses nanoseconds since 1970 as random seed generator, to approach random outcome.
	rand.Seed(time.Now().UnixNano())

	// Pseudo-random integer between 1 and 6.
	diceValue := rand.Intn(6) + 1

	return Modifier{Type: ModifierDice, Value: diceValue}
}

func ForestAttackerPenalty() Modifier {
	return Modifier{Type: ModifierForest, Value: -1}
}

func CastleAttackerPenalty() Modifier {
	return Modifier{Type: ModifierCastle, Value: -1}
}

func AttackAcrossWaterPenalty() Modifier {
	return Modifier{Type: ModifierWater, Value: -1}
}

func SurpriseAttackBonus() Modifier {
	return Modifier{Type: ModifierSurprise, Value: 1}
}

func SupportBonus(supportingPlayer string) Modifier {
	return Modifier{Type: ModifierSupport, Value: 1, SupportingPlayer: supportingPlayer}
}
