package board

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Returns modifiers (including dice roll) of defending unit in the area.
// Assumes that the area is not empty.
func (area Area) defenseModifiers() []Modifier {
	mods := []Modifier{}
	mods = appendUnitMod(mods, area.Unit.Type)
	mods = append(mods, Modifier{Type: ModifierDice, Value: rollDice()})
	return mods
}

// Returns modifiers (including dice roll) of move order attacking an area.
// Other parameters affect which modifiers are added:
// otherAttackers for whether there are other moves involved in this battle,
// borderBattle for whether this is a battle between two moves moving against each other,
// includeDefender for whether a potential defending unit in the area should be included.
func (move Order) attackModifiers(area Area, otherAttackers bool, borderBattle bool, includeDefender bool) []Modifier {
	mods := []Modifier{}

	neighbor, adjacent := area.GetNeighbor(move.From, move.Via)

	// Assumes danger zone checks have been made before battle,
	// and thus adds surprise modifier to attacker coming across such zones.
	if adjacent && neighbor.DangerZone != "" {
		mods = append(mods, Modifier{Type: ModifierSurprise, Value: 1})
	}

	// Terrain modifiers should be added if:
	// - Area is uncontrolled, and this unit is the only attacker.
	// - Destination is controlled and defended, and this is not a border conflict.
	if (!area.IsControlled() && !otherAttackers) ||
		(area.IsControlled() && !area.IsEmpty() && includeDefender && !borderBattle) {

		if area.Forest {
			mods = append(mods, Modifier{Type: ModifierForest, Value: -1})
		}

		if area.Castle {
			mods = append(mods, Modifier{Type: ModifierCastle, Value: -1})
		}

		// If origin area is not adjacent to destination, the move is transported and takes water penalty.
		// Moves across rivers or from sea to land also take this penalty.
		if !adjacent || neighbor.AcrossWater {
			mods = append(mods, Modifier{Type: ModifierWater, Value: -1})
		}
	}

	// Catapults get a bonus only in attacks on castle areas.
	if move.Unit.Type == UnitCatapult && area.Castle {
		mods = append(mods, Modifier{Type: ModifierUnit, Value: +1})
	} else {
		mods = appendUnitMod(mods, move.Unit.Type)
	}

	mods = append(mods, Modifier{Type: ModifierDice, Value: rollDice()})

	return mods
}

// Appends unit modifier to the given list if given unit type provides a modifier.
func appendUnitMod(mods []Modifier, unitType UnitType) []Modifier {
	switch unitType {
	case UnitFootman:
		return append(mods, Modifier{Type: ModifierUnit, Value: +1})
	default:
		return mods
	}
}

// Returns a pseudo-random integer between 1 and 6.
func rollDice() int {
	// Uses nanoseconds since 1970 as random seed generator, to approach random outcome.
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(6) + 1
}

type supportDeclaration struct {
	fromPlayer string
	toPlayer   string
}

// Calls support from support orders to the given area.
// Appends support modifiers to receiving players' results in the given map,
// but only if the result is tied to a move order to the area.
// Calls support to defender in the area if includeDefender is true.
func appendSupportMods(results map[string]Result, area Area, includeDefender bool, messenger Messenger) {
	supports := area.IncomingSupports
	supportCount := len(supports)
	supportReceiver := make(chan supportDeclaration, supportCount)
	var wg sync.WaitGroup
	wg.Add(supportCount)

	// Finds the moves going to this area.
	moves := []Order{}
	for _, result := range results {
		if result.DefenderArea != "" {
			continue
		}
		if result.Move.To == area.Name {
			moves = append(moves, result.Move)
		}
	}

	// Starts a goroutine to call support for each support order to the area.
	for _, support := range supports {
		go callSupport(support, area, moves, includeDefender, supportReceiver, &wg, messenger)
	}

	// Waits until all support calls are done, then closes the channel to range over it.
	wg.Wait()
	close(supportReceiver)

	for support := range supportReceiver {
		if support.toPlayer == "" {
			continue
		}

		result, isPlayer := results[support.toPlayer]
		if isPlayer {
			result.Parts = append(result.Parts, Modifier{Type: ModifierSupport, Value: 1, SupportingPlayer: support.fromPlayer})
			results[support.toPlayer] = result
		}
	}
}

// Finds out which player a given support order supports in a battle.
// Sends the resulting support declaration to the given supportReceiver, and decrements the wait group by 1.
//
// If the support order's player matches a player in the battle, support is automatically given to themselves.
// If support is not given to any player in the battle, the to field on the declaration is "".
//
// TODO: Implement asking player who to support if they are not involved themselves.
func callSupport(
	support Order,
	area Area,
	moves []Order,
	includeDefender bool,
	supportReceiver chan<- supportDeclaration,
	wg *sync.WaitGroup,
	messenger Messenger,
) {
	defer wg.Done()

	if includeDefender && !area.IsEmpty() && area.Unit.Player == support.Player {
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: support.Player}
		return
	}

	for _, move := range moves {
		if support.Player == move.Player {
			supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: support.Player}
			return
		}
	}

	battlers := make([]string, 0)
	for _, move := range moves {
		battlers = append(battlers, move.Player)
	}
	if includeDefender && !area.IsEmpty() {
		battlers = append(battlers, area.Unit.Player)
	}

	err := messenger.SendSupportRequest(support.Player, area.Name, battlers)
	if err != nil {
		log.Println(fmt.Errorf("failed to send support request: %w", err))
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: ""}
		return
	}

	supported, err := messenger.ReceiveSupport(support.Player, area.Name)
	if err != nil {
		log.Println(fmt.Errorf("failed to receive support declaration from player %s: %w", support.Player, err))
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: ""}
		return
	}

	supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: supported}
}

// Calculates totals for the given map of player IDs to results, and returns them as a list.
func calculateTotals(playerResults map[string]Result) []Result {
	results := make([]Result, 0)

	for _, result := range playerResults {
		total := 0
		for _, mod := range result.Parts {
			total += mod.Value
		}

		result.Total = total

		results = append(results, result)
	}

	return results
}
