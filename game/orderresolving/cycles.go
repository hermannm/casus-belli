package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/set"
)

// Resolves the board regions touched by the moves in the given cycle.
//
// Immediately resolves regions that do not require battle,
// and adds them to the given processed map.
//
// Adds embattled regions to the given processing map,
// and forwards them to appropriate battle calculators,
// which send results to the given battleReceiver.
func resolveCycle(
	cycle []gametypes.Order,
	board gametypes.Board,
	allowPlayerConflict bool,
	battleReceiver chan<- gametypes.Battle,
	processing set.Set[string],
	processed set.Set[string],
	messenger Messenger,
) {
	var battleRegions []gametypes.Region

	// First, resolves non-conflicting cycle moves.
	for _, move := range cycle {
		destination := board.Regions[move.Destination]

		if (destination.IsControlled() || destination.Sea) && len(destination.IncomingMoves) == 1 {
			succeedMove(move, board)
			processed.Add(destination.Name)
			continue
		}

		battleRegions = append(battleRegions, destination)
	}

	// Then resolves cycle moves that require battle.
	// Skips multiplayer battles if player conflicts are not allowed.
	for _, region := range battleRegions {
		if len(region.IncomingMoves) == 1 {
			go calculateSingleplayerBattle(
				region, region.IncomingMoves[0], battleReceiver, messenger,
			)
			processing.Add(region.Name)
		} else if allowPlayerConflict {
			go calculateMultiplayerBattle(region, false, battleReceiver, messenger)
			processing.Add(region.Name)
		} else {
			processed.Add(region.Name)
		}
	}
}

// Recursively finds a cycle of moves starting and ending with the given firstRegionName.
// Assumes that cycles of only 2 moves are already resolved.
// Returns the list of moves in the cycle, or nil if no cycle was found.
// Also returns whether any region in the cycle is attacked from outside the cycle.
func discoverCycle(
	firstRegionName string, order gametypes.Order, board gametypes.Board,
) (cycle []gametypes.Order, outsideAttackers bool) {
	if order.IsNone() || order.Type != gametypes.OrderMove {
		return nil, false
	}

	destination := board.Regions[order.Destination]

	// The cycle has outside attackers if more than just this order in the cycle is attacking the
	// destination.
	outsideAttackers = len(destination.IncomingMoves) > 1

	// The base case: the destination is the beginning of the cycle.
	if destination.Name == firstRegionName {
		return []gametypes.Order{order}, outsideAttackers
	}

	// If the base case is not yet reached, passes cycle discovery to the next order in the chain.
	continuedCycle, continuedOutsideAttackers := discoverCycle(
		firstRegionName, destination.Order, board,
	)
	if continuedCycle == nil {
		return nil, false
	} else {
		return append(continuedCycle, order), outsideAttackers || continuedOutsideAttackers
	}
}

// Checks if the given region is part of a two-way move cycle (moves moving against each other).
// Returns whether the region is aprt of a cycle, and if so, the second region in the cycle,
// as well as whether the two moves are from the same player.
func discoverTwoWayCycle(
	region1 gametypes.Region, board gametypes.Board,
) (isCycle bool, region2 gametypes.Region, samePlayer bool) {
	order1 := region1.Order
	if order1.Type != gametypes.OrderMove {
		return false, gametypes.Region{}, false
	}

	region2 = board.Regions[region1.Order.Destination]
	order2 := region2.Order
	if order2.Type != gametypes.OrderMove {
		return false, gametypes.Region{}, false
	}

	return order1.Origin == order2.Destination, region2, order1.Player == order2.Player
}
