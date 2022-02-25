package game

// Resolves the board areas touched by the moves in the given cycle.
//
// Immediately resolves areas that do not require battle,
// and adds them to the given processed map.
//
// Adds embattled areas to the given processing map,
// and forwards them to appropriate battle calculators,
// which send results to the given battleReceiver.
func (board Board) resolveCycle(
	cycle []Order,
	playerConflictsAllowed bool,
	battleReceiver chan<- Battle,
	processing map[string]struct{},
	processed map[string]struct{},
) {
	battleAreas := make([]Area, 0)

	// First, resolves non-conflicting cycle moves.
	for _, move := range cycle {
		to := board[move.To]

		if (to.IsControlled() || to.Sea) && len(to.IncomingMoves) == 1 {
			board.succeedMove(move)
			processed[to.Name] = struct{}{}
			continue
		}

		battleAreas = append(battleAreas, to)
	}

	// Then resolves cycle moves that require battle.
	// Skips multiplayer battles if player conflicts are not allowed.
	for _, area := range battleAreas {
		if len(area.IncomingMoves) == 1 {
			go area.calculateSingleplayerBattle(area.IncomingMoves[0], battleReceiver)
			processing[area.Name] = struct{}{}
		} else if playerConflictsAllowed {
			go area.calculateMultiplayerBattle(false, battleReceiver)
			processing[area.Name] = struct{}{}
		} else {
			processed[area.Name] = struct{}{}
		}
	}
}

// Recursively finds a cycle of moves starting and ending with the given firstAreaName.
// Assumes that cycles of only 2 moves are already resolved.
// Returns the list of moves in the cycle, or nil if no cycle was found.
// Also returns whether any area in the cycle is attacked from outside the cycle.
func (board Board) discoverCycle(order Order, firstAreaName string) (cycle []Order, outsideAttackers bool) {
	if order.IsNone() || order.Type != Move {
		return nil, false
	}

	to := board[order.To]

	// The cycle has outside attackers if more than just this order in the cycle is attacking the destination.
	outsideAttackers = len(to.IncomingMoves) > 1

	// The base case: the destination is the beginning of the cycle.
	if to.Name == firstAreaName {
		return []Order{order}, outsideAttackers
	}

	// If the base case is not yet reached, passes cycle discovery to the next order in the chain.
	continuedCycle, continuedOutsideAttackers := board.discoverCycle(to.Order, firstAreaName)
	if continuedCycle == nil {
		return nil, false
	} else {
		return append(continuedCycle, order), outsideAttackers || continuedOutsideAttackers
	}
}

// Checks if the given area is part of a two-way move cycle (moves moving against each other).
// Returns whether the area is aprt of a cycle, and if so, the second area in the cycle,
// as well as whether the two moves are from the same player.
func (board Board) discoverTwoWayCycle(area1 Area) (isCycle bool, area2 Area, samePlayer bool) {
	order1 := area1.Order
	if order1.Type != Move {
		return false, Area{}, false
	}

	area2 = board[area1.Order.To]
	order2 := area2.Order
	if order2.Type != Move {
		return false, Area{}, false
	}

	return order1.From == order2.To, area2, order1.Player == order2.Player
}
