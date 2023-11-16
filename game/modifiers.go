package game

import (
	"math/rand"
	"sync"

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

type supportDeclaration struct {
	from PlayerFaction
	to   PlayerFaction // Blank if nobody were supported.
}

// Calls support from support orders to the given region, and appends modifiers to the given map.
func (game *Game) appendSupportModifiers(
	results map[PlayerFaction]Result,
	region Region,
	includeDefender bool,
) {
	supports := region.IncomingSupports
	supportCount := len(supports)
	supportReceiver := make(chan supportDeclaration, supportCount)

	var waitGroup sync.WaitGroup
	waitGroup.Add(supportCount)

	incomingMoves := []Order{}
	for _, result := range results {
		if result.DefenderRegion != "" {
			continue
		}
		if result.Move.Destination == region.Name {
			incomingMoves = append(incomingMoves, result.Move)
		}
	}

	for _, support := range supports {
		go game.callSupport(
			support,
			region,
			incomingMoves,
			includeDefender,
			supportReceiver,
			&waitGroup,
		)
	}

	waitGroup.Wait()
	close(supportReceiver)

	for support := range supportReceiver {
		if support.to == "" {
			continue
		}

		result, isFaction := results[support.to]
		if isFaction {
			result.Parts = append(
				result.Parts,
				Modifier{Type: ModifierSupport, Value: 1, SupportingFaction: support.from},
			)
			results[support.to] = result
		}
	}
}

func (game *Game) callSupport(
	support Order,
	region Region,
	moves []Order,
	includeDefender bool,
	supportReceiver chan<- supportDeclaration,
	waitGroup *sync.WaitGroup,
) {
	defer waitGroup.Done()

	if includeDefender && !region.isEmpty() && region.Unit.Faction == support.Faction {
		supportReceiver <- supportDeclaration{from: support.Faction, to: support.Faction}
		return
	}

	for _, move := range moves {
		if support.Faction == move.Faction {
			supportReceiver <- supportDeclaration{from: support.Faction, to: support.Faction}
			return
		}
	}

	battlers := make([]PlayerFaction, 0, len(moves)+1)
	for _, move := range moves {
		battlers = append(battlers, move.Faction)
	}
	if includeDefender && !region.isEmpty() {
		battlers = append(battlers, region.Unit.Faction)
	}

	if err := game.messenger.SendSupportRequest(
		support.Faction,
		support.Origin,
		region.Name,
		battlers,
	); err != nil {
		game.log.ErrorCause(err, "failed to send support request")
		supportReceiver <- supportDeclaration{from: support.Faction, to: ""}
		return
	}

	supported, err := game.messenger.AwaitSupport(support.Faction, support.Origin, region.Name)
	if err != nil {
		game.log.ErrorCausef(
			err,
			"failed to receive support declaration from faction '%s'",
			support.Faction,
		)
		supportReceiver <- supportDeclaration{from: support.Faction, to: ""}
		return
	}

	supportReceiver <- supportDeclaration{from: support.Faction, to: supported}
}
