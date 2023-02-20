package gameboard

// Resolves the board regions touched by the moves in the given cycle.
//
// Immediately resolves regions that do not require battle,
// and adds them to the given processed map.
//
// Adds embattled regions to the given processing map,
// and forwards them to appropriate battle calculators,
// which send results to the given battleReceiver.
func (board Board) resolveCycle(
	cycle []Order,
	allowPlayerConflict bool,
	battleReceiver chan<- Battle,
	processing map[string]struct{},
	processed map[string]struct{},
	messenger Messenger,
) {
	battleRegions := make([]Region, 0)

	// First, resolves non-conflicting cycle moves.
	for _, move := range cycle {
		to := board.Regions[move.To]

		if (to.IsControlled() || to.Sea) && len(to.IncomingMoves) == 1 {
			board.succeedMove(move)
			processed[to.Name] = struct{}{}
			continue
		}

		battleRegions = append(battleRegions, to)
	}

	// Then resolves cycle moves that require battle.
	// Skips multiplayer battles if player conflicts are not allowed.
	for _, region := range battleRegions {
		if len(region.IncomingMoves) == 1 {
			go region.calculateSingleplayerBattle(
				region.IncomingMoves[0],
				battleReceiver,
				messenger,
			)
			processing[region.Name] = struct{}{}
		} else if allowPlayerConflict {
			go region.calculateMultiplayerBattle(false, battleReceiver, messenger)
			processing[region.Name] = struct{}{}
		} else {
			processed[region.Name] = struct{}{}
		}
	}
}

// Recursively finds a cycle of moves starting and ending with the given firstRegionName.
// Assumes that cycles of only 2 moves are already resolved.
// Returns the list of moves in the cycle, or nil if no cycle was found.
// Also returns whether any region in the cycle is attacked from outside the cycle.
func (board Board) discoverCycle(order Order, firstRegionName string) (cycle []Order, outsideAttackers bool) {
	if order.IsNone() || order.Type != OrderMove {
		return nil, false
	}

	to := board.Regions[order.To]

	// The cycle has outside attackers if more than just this order in the cycle is attacking the destination.
	outsideAttackers = len(to.IncomingMoves) > 1

	// The base case: the destination is the beginning of the cycle.
	if to.Name == firstRegionName {
		return []Order{order}, outsideAttackers
	}

	// If the base case is not yet reached, passes cycle discovery to the next order in the chain.
	continuedCycle, continuedOutsideAttackers := board.discoverCycle(to.Order, firstRegionName)
	if continuedCycle == nil {
		return nil, false
	} else {
		return append(continuedCycle, order), outsideAttackers || continuedOutsideAttackers
	}
}

// Checks if the given region is part of a two-way move cycle (moves moving against each other).
// Returns whether the region is aprt of a cycle, and if so, the second region in the cycle,
// as well as whether the two moves are from the same player.
func (board Board) discoverTwoWayCycle(region1 Region) (isCycle bool, region2 Region, samePlayer bool) {
	order1 := region1.Order
	if order1.Type != OrderMove {
		return false, Region{}, false
	}

	region2 = board.Regions[region1.Order.To]
	order2 := region2.Order
	if order2.Type != OrderMove {
		return false, Region{}, false
	}

	return order1.From == order2.To, region2, order1.Player == order2.Player
}
