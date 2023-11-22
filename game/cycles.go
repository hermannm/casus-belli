package game

// Recursively finds a cycle of moves starting and ending with the given firstRegionName.
func (board Board) discoverCycle(
	firstRegionName RegionName,
	order Order,
) (cycle []Order, hasOutsideAttackers bool) {
	if order.isNone() || order.Type != OrderMove {
		return nil, false
	}

	destination := board[order.Destination]

	// The cycle has outside attackers if more than just this order in the cycle is attacking the
	// destination.
	hasOutsideAttackers = len(destination.incomingMoves) > 1

	// The base case: the destination is the beginning of the cycle.
	if destination.Name == firstRegionName {
		return []Order{order}, hasOutsideAttackers
	}

	// If the base case is not yet reached, passes cycle discovery to the next order in the chain.
	continuedCycle, continuedOutsideAttackers := board.discoverCycle(
		firstRegionName,
		destination.order,
	)
	if continuedCycle == nil {
		return nil, false
	} else {
		return append(continuedCycle, order), hasOutsideAttackers || continuedOutsideAttackers
	}
}

// Checks if the given region is part of a two-way move cycle (moves moving against each other).
func (board Board) discoverTwoWayCycle(
	region1 *Region,
) (isCycle bool, region2 *Region, sameFaction bool) {
	order1 := region1.order
	if order1.Type != OrderMove {
		return false, nil, false
	}

	region2 = board[region1.order.Destination]
	order2 := region2.order
	if order2.Type != OrderMove {
		return false, nil, false
	}

	return order1.Origin == order2.Destination, region2, order1.Faction == order2.Faction
}

func (game *Game) resolveCycle(cycle []Order) {
	regions := make([]*Region, len(cycle))

	for i, move := range cycle {
		destination := game.board[move.Destination]
		destination.removeUnit()
		destination.order = Order{}
		destination.partOfCycle = true
		regions[i] = destination
	}

	for _, region := range regions {
		if len(region.incomingMoves) == 1 {
			if region.controlled() || region.Sea {
				game.board.succeedMove(region.incomingMoves[0])
			} else {
				game.calculateSingleplayerBattle(region)
			}
		} else {
			game.calculateMultiplayerBattle(region)
		}
	}
}
