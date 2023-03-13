package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/set"
)

type ResolverState struct {
	allowPlayerConflict bool
	resolvedBattles     []gametypes.Battle
	battleReceiver      chan gametypes.Battle
	processing          set.Set[string]
	processed           set.Set[string]
	retreats            map[string]gametypes.Order
}

func newResolverState(allowPlayerConflict bool) ResolverState {
	return ResolverState{
		allowPlayerConflict: allowPlayerConflict,
		resolvedBattles:     nil,
		battleReceiver:      make(chan gametypes.Battle),
		processing:          set.New[string](),
		processed:           set.New[string](),
		retreats:            make(map[string]gametypes.Order),
	}
}

// Resolves moves on the board. Returns any resulting battles.
// Only resolves battles between players if allowPlayerConflict is true.
func resolveMoves(
	board gametypes.Board, allowPlayerConflict bool, messenger Messenger,
) []gametypes.Battle {
	resolverState := newResolverState(allowPlayerConflict)

OuterLoop:
	for {
		select {
		case battle := <-resolverState.battleReceiver:
			resolverState.resolvedBattles = append(resolverState.resolvedBattles, battle)
			messenger.SendBattleResults([]gametypes.Battle{battle})

			newRetreats := resolveBattle(battle, board)
			for _, retreat := range newRetreats {
				resolverState.retreats[retreat.Origin] = retreat
			}

			for _, region := range battle.RegionNames() {
				resolverState.processing.Remove(region)
			}
		default:
			for _, region := range board.Regions {
				resolveRegionMoves(region, board, &resolverState, messenger)
			}

			if resolverState.processing.IsEmpty() && len(resolverState.retreats) == 0 {
				break OuterLoop
			}
		}
	}

	return resolverState.resolvedBattles
}

// Resolves moves to the given region on the board.
//
// Regions that do not require battle are immediately resolved. Regions that do require battle are
// forwarded to appropriate battle calculation functions, which send results on the battle receiver
// channel in the given ResolverState.
//
// Skips regions that have outgoing moves, unless they are part of a move cycle.
// If resolverState.allowPlayerConflict is false, skips regions that require battle between players.
func resolveRegionMoves(
	region gametypes.Region,
	board gametypes.Board,
	resolverState *ResolverState,
	messenger Messenger,
) {
	retreat, hasRetreat := resolverState.retreats[region.Name]

	// Skips the region if it has already been processed.
	if (resolverState.processed.Contains(region.Name) && !hasRetreat) ||
		(resolverState.processing.Contains(region.Name)) {
		return
	}

	// Resolves incoming moves that require transport.
	for _, move := range region.IncomingMoves {
		transportMustWait := resolveTransport(move, board, resolverState, messenger)
		if transportMustWait {
			return
		}
	}

	// Resolves retreats if region has no attackers.
	if !region.IsAttacked() {
		if hasRetreat && region.IsEmpty() {
			region.Unit = retreat.Unit
			board.Regions[region.Name] = region
			delete(resolverState.retreats, region.Name)
		}

		resolverState.processed.Add(region.Name)
		return
	}

	// Finds out if the region part of a cycle (moves in a circle).
	twoWayCycle, region2, samePlayer := discoverTwoWayCycle(region, board)
	if twoWayCycle && samePlayer {
		// If both moves are by the same player, removes the units from their origin regions,
		// as they may not be allowed to retreat if their origin region is taken.
		for _, cycleRegion := range [2]gametypes.Region{region, region2} {
			cycleRegion.Unit = gametypes.Unit{}
			cycleRegion.Order = gametypes.Order{}
			board.Regions[cycleRegion.Name] = cycleRegion
		}
	} else if twoWayCycle {
		// If the moves are from different players, they battle in the middle.
		go calculateBorderBattle(region, region2, resolverState.battleReceiver, messenger)
		resolverState.processing.Add(region.Name, region2.Name)
		return
	} else if cycle, _ := discoverCycle(region.Name, region.Order, board); cycle != nil {
		// If there is a cycle longer than 2 moves, forwards the resolving to resolveCycle.
		resolveCycle(cycle, board, resolverState, messenger)
		return
	}

	// A single move to an empty region is either an autosuccess, or a singleplayer battle.
	if len(region.IncomingMoves) == 1 && region.IsEmpty() {
		move := region.IncomingMoves[0]

		if region.IsControlled() || region.Sea {
			succeedMove(move, board)
			resolverState.processed.Add(region.Name)
			return
		}

		go calculateSingleplayerBattle(region, move, resolverState.battleReceiver, messenger)
		resolverState.processing.Add(region.Name)
		return
	}

	// If the destination region has an outgoing move order, that must be resolved first.
	if region.Order.Type == gametypes.OrderMove {
		return
	}

	// If the function has not returned yet, then it must be a multiplayer battle.
	go calculateMultiplayerBattle(
		region, !region.IsEmpty(), resolverState.battleReceiver, messenger,
	)
	resolverState.processing.Add(region.Name)
}
