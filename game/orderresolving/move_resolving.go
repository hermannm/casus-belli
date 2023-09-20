package orderresolving

import (
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/set"
)

type MoveResolver struct {
	resolvingRegions   set.Set[string]
	resolvedRegions    set.Set[string]
	resolvedTransports set.Set[string]
	resolvedBattles    []gametypes.Battle
	battleReceiver     chan gametypes.Battle
	retreats           map[string]gametypes.Order
	secondHorseMoves   []gametypes.Order
}

func newMoveResolver() MoveResolver {
	return MoveResolver{
		resolvingRegions:   set.New[string](),
		resolvedRegions:    set.New[string](),
		resolvedTransports: set.New[string](),
		resolvedBattles:    nil,
		battleReceiver:     make(chan gametypes.Battle),
		retreats:           make(map[string]gametypes.Order),
		secondHorseMoves:   nil,
	}
}

func (resolver *MoveResolver) resolveMoves(board gametypes.Board, messenger Messenger) {
OuterLoop:
	for {
		select {
		case battle := <-resolver.battleReceiver:
			resolver.resolveBattle(battle, board)
			messenger.SendBattleResults([]gametypes.Battle{battle})
		default:
			for _, region := range board.Regions {
				resolver.resolveRegionMoves(region, board, messenger)
			}

			if resolver.resolvingRegions.IsEmpty() && len(resolver.retreats) == 0 {
				break OuterLoop
			}
		}
	}
}

// Immediately resolves region if it does not require battle. If it does require battle, forwards it
// to appropriate battle calculation functions, which send results to resolver.battleReceiver.
// Skips region if it depends on other moves to resolve first.
func (resolver *MoveResolver) resolveRegionMoves(
	region gametypes.Region,
	board gametypes.Board,
	messenger Messenger,
) {
	retreat, hasRetreat := resolver.retreats[region.Name]

	// Skips the region if it has already been processed
	if (resolver.resolvedRegions.Contains(region.Name) && !hasRetreat) ||
		(resolver.resolvingRegions.Contains(region.Name)) {
		return
	}

	// Resolves incoming moves that require transport
	if !resolver.resolvedTransports.Contains(region.Name) {
		resolver.resolvedTransports.Add(region.Name)

		for _, move := range region.IncomingMoves {
			transportMustWait := resolver.resolveTransport(move, board, messenger)
			if transportMustWait {
				return
			}
		}
	}

	// Resolves retreats if region has no attackers
	if !region.IsAttacked() {
		if hasRetreat && region.IsEmpty() {
			region.Unit = retreat.Unit
			board.Regions[region.Name] = region
			delete(resolver.retreats, region.Name)
		}

		resolver.resolvedRegions.Add(region.Name)
		return
	}

	// Finds out if the region is part of a cycle (moves in a circle)
	twoWayCycle, region2, samePlayer := discoverTwoWayCycle(region, board)
	if twoWayCycle && samePlayer {
		// If both moves are by the same player, removes the units from their origin regions,
		// as they may not be allowed to retreat if their origin region is taken
		for _, cycleRegion := range [2]gametypes.Region{region, region2} {
			cycleRegion.Unit = gametypes.Unit{}
			cycleRegion.Order = gametypes.Order{}
			board.Regions[cycleRegion.Name] = cycleRegion
		}
	} else if twoWayCycle {
		// If the moves are from different players, they battle in the middle
		go calculateBorderBattle(region, region2, resolver.battleReceiver, messenger)
		resolver.resolvingRegions.Add(region.Name, region2.Name)
		return
	} else if cycle, _ := discoverCycle(region.Name, region.Order, board); cycle != nil {
		// If there is a cycle longer than 2 moves, forwards the resolving to resolveCycle
		resolver.resolveCycle(cycle, board, messenger)
		return
	}

	// A single move to an empty region is either an autosuccess, or a singleplayer battle
	if len(region.IncomingMoves) == 1 && region.IsEmpty() {
		move := region.IncomingMoves[0]

		if region.IsControlled() || region.IsSea {
			resolver.succeedMove(move, board)
			return
		}

		go calculateSingleplayerBattle(region, move, resolver.battleReceiver, messenger)
		resolver.resolvingRegions.Add(region.Name)
		return
	}

	// If the destination region has an outgoing move order, that must be resolved first
	if region.Order.Type == gametypes.OrderMove {
		return
	}

	// If the function has not returned yet, then it must be a multiplayer battle
	go calculateMultiplayerBattle(region, !region.IsEmpty(), resolver.battleReceiver, messenger)
	resolver.resolvingRegions.Add(region.Name)
}

func (resolver *MoveResolver) resolveTransport(
	move gametypes.Order,
	board gametypes.Board,
	messenger Messenger,
) (transportMustWait bool) {
	// If the move is between two adjacent regions, then it does not need transport
	if board.Regions[move.Destination].HasNeighbor(move.Origin) {
		return false
	}

	canTransport, transportAttacked, dangerZones := board.FindTransportPath(
		move.Origin,
		move.Destination,
	)

	if !canTransport {
		board.RemoveOrder(move)
		return false
	}

	if transportAttacked {
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

func (resolver *MoveResolver) resolveCycle(
	cycle []gametypes.Order,
	board gametypes.Board,
	messenger Messenger,
) {
	var battleRegions []gametypes.Region

	// First resolves non-conflicting cycle moves
	for _, move := range cycle {
		destination := board.Regions[move.Destination]

		if (destination.IsControlled() || destination.IsSea) &&
			len(destination.IncomingMoves) == 1 {
			resolver.succeedMove(move, board)
			continue
		}

		battleRegions = append(battleRegions, destination)
	}

	// Then resolves cycle moves that require battle
	for _, region := range battleRegions {
		if len(region.IncomingMoves) == 1 {
			go calculateSingleplayerBattle(
				region,
				region.IncomingMoves[0],
				resolver.battleReceiver,
				messenger,
			)
			resolver.resolvingRegions.Add(region.Name)
		} else {
			go calculateMultiplayerBattle(region, false, resolver.battleReceiver, messenger)
			resolver.resolvingRegions.Add(region.Name)
		}
	}
}

func (resolver *MoveResolver) succeedMove(move gametypes.Order, board gametypes.Board) {
	destination := board.Regions[move.Destination]

	destination.Unit = move.Unit
	destination.Order = gametypes.Order{}
	if !destination.IsSea {
		destination.ControllingPlayer = move.Player
	}

	board.Regions[move.Destination] = destination

	board.RemoveUnit(move.Unit, move.Origin)
	board.RemoveOrder(move)

	resolver.resolvedRegions.Add(move.Destination)

	if secondHorseMove, hasSecondHorseMove := move.TryGetSecondHorseMove(); hasSecondHorseMove {
		resolver.secondHorseMoves = append(resolver.secondHorseMoves, secondHorseMove)
	}
}

func (resolver *MoveResolver) addSecondHorseMoves(board gametypes.Board) {
	for _, secondHorseMove := range resolver.secondHorseMoves {
		board.AddOrder(secondHorseMove)
		resolver.resolvedRegions.Remove(secondHorseMove.Destination)
	}

	resolver.secondHorseMoves = nil
}
