package orderresolving

import (
	"fmt"
	"log"
	"sync"

	"hermannm.dev/bfh-server/game/gametypes"
)

// Resolves moves to the given region on the board.
// Assumes that the region has incoming moves (moveCount > 0).
//
// Immediately resolves regions that do not require battle, and adds them to the given processed
// map.
//
// Adds embattled regions to the given processing map, and forwards them to appropriate battle
// calculation functions, which send results to the given battleReceiver.
//
// Skips regions that have outgoing moves, unless they are part of a move cycle.
// If allowPlayerConflict is false, skips regions that require battle between players.
func resolveRegionMoves(
	region gametypes.Region,
	board gametypes.Board,
	moveCount int,
	allowPlayerConflict bool,
	battleReceiver chan gametypes.Battle,
	processing map[string]struct{},
	processed map[string]struct{},
	messenger Messenger,
) {
	// Finds out if the move is part of a two-way cycle (moves moving against each other),
	// and resolves it.
	twoWayCycle, region2, samePlayer := discoverTwoWayCycle(region, board)
	if twoWayCycle {
		if samePlayer {
			// If both moves are by the same player, removes the units from their origin regions,
			// as they may not be allowed to retreat if their origin region is taken.
			for _, cycleRegion := range [2]gametypes.Region{region, region2} {
				cycleRegion.Unit = gametypes.Unit{}
				cycleRegion.Order = gametypes.Order{}
				board.Regions[cycleRegion.Name] = cycleRegion
			}
		} else {
			// If the moves are from different players, they battle in the middle.
			go calculateBorderBattle(region, region2, battleReceiver, messenger)
			processing[region.Name], processing[region2.Name] = struct{}{}, struct{}{}
			return
		}
	} else {
		// If there is a cycle longer than 2 moves, forwards the resolving to 'resolveCycle'.
		cycle, _ := discoverCycle(region.Name, region.Order, board)
		if cycle != nil {
			resolveCycle(
				cycle,
				board,
				allowPlayerConflict,
				battleReceiver,
				processing,
				processed,
				messenger,
			)
			return
		}
	}

	// Empty regions with only a single incoming move are either auto-successes or a singleplayer
	// battle.
	if moveCount == 1 && region.IsEmpty() {
		move := region.IncomingMoves[0]

		if region.IsControlled() || region.Sea {
			succeedMove(move, board)
			processed[region.Name] = struct{}{}
			return
		}

		go calculateSingleplayerBattle(region, move, battleReceiver, messenger)
		processing[region.Name] = struct{}{}
		return
	}

	// If the destination region has an outgoing move order, that must be resolved first.
	if region.Order.Type == gametypes.OrderMove {
		return
	}

	// If the function has not returned yet, then it must be a multiplayer battle.
	go calculateMultiplayerBattle(region, !region.IsEmpty(), battleReceiver, messenger)
	processing[region.Name] = struct{}{}
}

// Calculates battle between a single attacker and an unconquered region.
// Sends the resulting battle to the given battleReceiver.
func calculateSingleplayerBattle(
	region gametypes.Region,
	move gametypes.Order,
	battleReceiver chan<- gametypes.Battle,
	messenger Messenger,
) {
	playerResults := map[string]gametypes.Result{
		move.Player: {Parts: attackModifiers(move, region, false, false, true), Move: move},
	}

	appendSupportMods(playerResults, region, false, messenger)

	battleReceiver <- gametypes.Battle{Results: calculateTotals(playerResults)}
}

// Calculates battle when attacked region is defended or has multiple attackers.
// Takes in parameter for whether to account for defender in battle (most often true).
// Sends the resulting battle to the given battleReceiver.
func calculateMultiplayerBattle(
	region gametypes.Region,
	includeDefender bool,
	battleReceiver chan<- gametypes.Battle,
	messenger Messenger,
) {
	playerResults := make(map[string]gametypes.Result)

	for _, move := range region.IncomingMoves {
		playerResults[move.Player] = gametypes.Result{
			Parts: attackModifiers(move, region, true, false, includeDefender),
			Move:  move,
		}
	}

	if !region.IsEmpty() && includeDefender {
		playerResults[region.Unit.Player] = gametypes.Result{
			Parts:          defenseModifiers(region),
			DefenderRegion: region.Name,
		}
	}

	appendSupportMods(playerResults, region, includeDefender, messenger)

	battleReceiver <- gametypes.Battle{Results: calculateTotals(playerResults)}
}

// Calculates battle when units from two regions attack each other simultaneously.
// Sends the resulting battle to the given battleReceiver.
func calculateBorderBattle(
	region1 gametypes.Region,
	region2 gametypes.Region,
	battleReceiver chan<- gametypes.Battle,
	messenger Messenger,
) {
	move1 := region1.Order
	move2 := region2.Order
	playerResults := map[string]gametypes.Result{
		move1.Player: {Parts: attackModifiers(move1, region2, true, true, false), Move: move1},
		move2.Player: {Parts: attackModifiers(move2, region1, true, true, false), Move: move2},
	}

	appendSupportMods(playerResults, region2, false, messenger)
	appendSupportMods(playerResults, region1, false, messenger)

	battleReceiver <- gametypes.Battle{Results: calculateTotals(playerResults)}
}

// Returns modifiers (including dice roll) of defending unit in the region.
// Assumes that the region is not empty.
func defenseModifiers(region gametypes.Region) []gametypes.Modifier {
	modifiers := []gametypes.Modifier{gametypes.RollDiceBonus()}

	if unitModifier, hasModifier := region.Unit.BattleModifier(false); hasModifier {
		modifiers = append(modifiers, unitModifier)
	}

	return modifiers
}

// Returns modifiers (including dice roll) of move order attacking a region.
// Other parameters affect which modifiers are added:
// otherAttackers for whether there are other moves involved in this battle,
// borderBattle for whether this is a battle between two moves moving against each other,
// includeDefender for whether a potential defending unit in the region should be included.
func attackModifiers(
	move gametypes.Order,
	region gametypes.Region,
	otherAttackers bool,
	borderBattle bool,
	includeDefender bool,
) []gametypes.Modifier {
	mods := []gametypes.Modifier{}

	neighbor, adjacent := region.GetNeighbor(move.From, move.Via)

	// Assumes danger zone checks have been made before battle,
	// and thus adds surprise modifier to attacker coming across such zones.
	if adjacent && neighbor.DangerZone != "" {
		mods = append(mods, gametypes.SurpriseAttackBonus())
	}

	// Terrain modifiers should be added if:
	// - Region is uncontrolled, and this unit is the only attacker.
	// - Destination is controlled and defended, and this is not a border conflict.
	if (!region.IsControlled() && !otherAttackers) ||
		(region.IsControlled() && !region.IsEmpty() && includeDefender && !borderBattle) {

		if region.Forest {
			mods = append(mods, gametypes.ForestAttackerPenalty())
		}

		if region.Castle {
			mods = append(mods, gametypes.CastleAttackerPenalty())
		}

		// If origin region is not adjacent to destination, the move is transported and takes water
		// penalty. Moves across rivers or from sea to land also take this penalty.
		if !adjacent || neighbor.AcrossWater {
			mods = append(mods, gametypes.AttackAcrossWaterPenalty())
		}
	}

	if unitModifier, hasModifier := region.Unit.BattleModifier(region.Castle); hasModifier {
		mods = append(mods, unitModifier)
	}

	mods = append(mods, gametypes.RollDiceBonus())

	return mods
}

type supportDeclaration struct {
	fromPlayer string
	toPlayer   string
}

// Calls support from support orders to the given region.
// Appends support modifiers to receiving players' results in the given map,
// but only if the result is tied to a move order to the region.
// Calls support to defender in the region if includeDefender is true.
func appendSupportMods(
	results map[string]gametypes.Result,
	region gametypes.Region,
	includeDefender bool,
	messenger Messenger,
) {
	supports := region.IncomingSupports
	supportCount := len(supports)
	supportReceiver := make(chan supportDeclaration, supportCount)
	var wg sync.WaitGroup
	wg.Add(supportCount)

	// Finds the moves going to this region.
	moves := []gametypes.Order{}
	for _, result := range results {
		if result.DefenderRegion != "" {
			continue
		}
		if result.Move.To == region.Name {
			moves = append(moves, result.Move)
		}
	}

	// Starts a goroutine to call support for each support order to the region.
	for _, support := range supports {
		go callSupport(support, region, moves, includeDefender, supportReceiver, &wg, messenger)
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
			result.Parts = append(result.Parts, gametypes.SupportBonus(support.fromPlayer))
			results[support.toPlayer] = result
		}
	}
}

// Finds out which player a given support order supports in a battle. Sends the resulting support
// declaration to the given supportReceiver, and decrements the wait group by 1.
//
// If the support order's player matches a player in the battle, support is automatically given to
// themselves.
// If support is not given to any player in the battle, the to field on the declaration is "".
func callSupport(
	support gametypes.Order,
	region gametypes.Region,
	moves []gametypes.Order,
	includeDefender bool,
	supportReceiver chan<- supportDeclaration,
	wg *sync.WaitGroup,
	messenger Messenger,
) {
	defer wg.Done()

	if includeDefender && !region.IsEmpty() && region.Unit.Player == support.Player {
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: support.Player}
		return
	}

	for _, move := range moves {
		if support.Player == move.Player {
			supportReceiver <- supportDeclaration{
				fromPlayer: support.Player,
				toPlayer:   support.Player,
			}
			return
		}
	}

	battlers := make([]string, 0)
	for _, move := range moves {
		battlers = append(battlers, move.Player)
	}
	if includeDefender && !region.IsEmpty() {
		battlers = append(battlers, region.Unit.Player)
	}

	err := messenger.SendSupportRequest(support.Player, region.Name, battlers)
	if err != nil {
		log.Println(fmt.Errorf("failed to send support request: %w", err))
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: ""}
		return
	}

	supported, err := messenger.ReceiveSupport(support.Player, region.Name)
	if err != nil {
		log.Println(fmt.Errorf(
			"failed to receive support declaration from player %s: %w",
			support.Player,
			err,
		))
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: ""}
		return
	}

	supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: supported}
}

// Calculates totals for the given map of player IDs to results, and returns them as a list.
func calculateTotals(playerResults map[string]gametypes.Result) []gametypes.Result {
	results := make([]gametypes.Result, 0)

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
