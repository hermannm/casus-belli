package orderresolving

import (
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/set"
)

type MoveResolver struct {
	allowPlayerConflict bool
	resolvedBattles     []gametypes.Battle
	battleReceiver      chan gametypes.Battle
	processing          set.Set[string]
	processed           set.Set[string]
	retreats            map[string]gametypes.Order
}

func newMoveResolver() MoveResolver {
	return MoveResolver{
		allowPlayerConflict: false,
		resolvedBattles:     nil,
		battleReceiver:      make(chan gametypes.Battle),
		processing:          set.New[string](),
		processed:           set.New[string](),
		retreats:            make(map[string]gametypes.Order),
	}
}

// Resolves moves on the board. Resulting battles are added to resolver.resolvedBattles.
// Only resolves battles between players if resolver.allowPlayerConflict is true.
func (resolver *MoveResolver) resolveMoves(board gametypes.Board, messenger Messenger) {
OuterLoop:
	for {
		select {
		case battle := <-resolver.battleReceiver:
			resolver.resolvedBattles = append(resolver.resolvedBattles, battle)
			messenger.SendBattleResults([]gametypes.Battle{battle})

			newRetreats := resolveBattle(battle, board)
			for _, retreat := range newRetreats {
				resolver.retreats[retreat.Origin] = retreat
			}

			for _, region := range battle.RegionNames() {
				resolver.processing.Remove(region)
			}
		default:
			for _, region := range board.Regions {
				resolver.resolveRegionMoves(region, board, messenger)
			}

			if resolver.processing.IsEmpty() && len(resolver.retreats) == 0 {
				break OuterLoop
			}
		}
	}
}

// Resolves moves to the given region on the board.
//
// Regions that do not require battle are immediately resolved. Regions that do require battle are
// forwarded to appropriate battle calculation functions, which send results on the battle receiver
// channel in the given ResolverState.
//
// Skips regions that have outgoing moves, unless they are part of a move cycle.
// If resolverState.allowPlayerConflict is false, skips regions that require battle between players.
func (resolver *MoveResolver) resolveRegionMoves(
	region gametypes.Region, board gametypes.Board, messenger Messenger,
) {
	retreat, hasRetreat := resolver.retreats[region.Name]

	// Skips the region if it has already been processed.
	if (resolver.processed.Contains(region.Name) && !hasRetreat) ||
		(resolver.processing.Contains(region.Name)) {
		return
	}

	// Resolves incoming moves that require transport.
	for _, move := range region.IncomingMoves {
		transportMustWait := resolver.resolveTransport(move, board, messenger)
		if transportMustWait {
			return
		}
	}

	// Resolves retreats if region has no attackers.
	if !region.IsAttacked() {
		if hasRetreat && region.IsEmpty() {
			region.Unit = retreat.Unit
			board.Regions[region.Name] = region
			delete(resolver.retreats, region.Name)
		}

		resolver.processed.Add(region.Name)
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
		go calculateBorderBattle(region, region2, resolver.battleReceiver, messenger)
		resolver.processing.Add(region.Name, region2.Name)
		return
	} else if cycle, _ := discoverCycle(region.Name, region.Order, board); cycle != nil {
		// If there is a cycle longer than 2 moves, forwards the resolving to resolveCycle.
		resolver.resolveCycle(cycle, board, messenger)
		return
	}

	// A single move to an empty region is either an autosuccess, or a singleplayer battle.
	if len(region.IncomingMoves) == 1 && region.IsEmpty() {
		move := region.IncomingMoves[0]

		if region.IsControlled() || region.Sea {
			succeedMove(move, board)
			resolver.processed.Add(region.Name)
			return
		}

		go calculateSingleplayerBattle(region, move, resolver.battleReceiver, messenger)
		resolver.processing.Add(region.Name)
		return
	}

	// If the destination region has an outgoing move order, that must be resolved first.
	if region.Order.Type == gametypes.OrderMove {
		return
	}

	// If the function has not returned yet, then it must be a multiplayer battle.
	go calculateMultiplayerBattle(
		region, !region.IsEmpty(), resolver.battleReceiver, messenger,
	)
	resolver.processing.Add(region.Name)
}

// Resolves transport of the given move to its destination, if it requires transport.
// If the transport depends on other orders to resolve first, returns transportMustWait=true.
func (resolver *MoveResolver) resolveTransport(
	move gametypes.Order, board gametypes.Board, messenger Messenger,
) (transportMustWait bool) {
	// If the move is between two adjacent regions, then it does not need transport.
	if board.Regions[move.Destination].HasNeighbor(move.Origin) {
		return false
	}

	canTransport, transportAttacked, dangerZones := board.FindTransportPath(move.Origin, move.Destination)

	if !canTransport {
		board.RemoveOrder(move)
		return false
	}

	if transportAttacked {
		if resolver.allowPlayerConflict {
			resolver.processed.Add(move.Destination)
		}

		return true
	}

	if len(dangerZones) > 0 {
		survived, dangerZoneBattles := crossDangerZones(move, dangerZones)

		if !survived {
			board.RemoveOrder(move)
		}

		resolver.resolvedBattles = append(resolver.resolvedBattles, dangerZoneBattles...)
		if err := messenger.SendBattleResults(dangerZoneBattles); err != nil {
			log.Println(err)
		}

		return false
	}

	return false
}

// Resolves the board regions touched by the moves in the given cycle.
//
// Immediately resolves regions that do not require battle,
// and adds them to the given processed map.
//
// Adds embattled regions to the given processing map,
// and forwards them to appropriate battle calculators,
// which send results to the given battleReceiver.
func (resolver *MoveResolver) resolveCycle(
	cycle []gametypes.Order, board gametypes.Board, messenger Messenger,
) {
	var battleRegions []gametypes.Region

	// First, resolves non-conflicting cycle moves.
	for _, move := range cycle {
		destination := board.Regions[move.Destination]

		if (destination.IsControlled() || destination.Sea) && len(destination.IncomingMoves) == 1 {
			succeedMove(move, board)
			resolver.processed.Add(destination.Name)
			continue
		}

		battleRegions = append(battleRegions, destination)
	}

	// Then resolves cycle moves that require battle.
	// Skips multiplayer battles if player conflicts are not allowed.
	for _, region := range battleRegions {
		if len(region.IncomingMoves) == 1 {
			go calculateSingleplayerBattle(
				region, region.IncomingMoves[0], resolver.battleReceiver, messenger,
			)
			resolver.processing.Add(region.Name)
		} else if resolver.allowPlayerConflict {
			go calculateMultiplayerBattle(region, false, resolver.battleReceiver, messenger)
			resolver.processing.Add(region.Name)
		} else {
			resolver.processed.Add(region.Name)
		}
	}
}
