package game

// Recursively finds a cycle of move orders through regions starting and ending with the given
// firstRegionName.
func (board Board) discoverCycle(firstRegionName RegionName, order Order) (cycle []*Region) {
	if order.Type != OrderMove {
		return nil
	}

	destination := board[order.Destination]

	// The base case: the destination is the beginning of the cycle.
	if destination.Name == firstRegionName {
		return []*Region{destination}
	}

	// If the base case is not yet reached, passes cycle discovery to the next order in the chain.
	cycle = board.discoverCycle(firstRegionName, destination.order)
	if cycle == nil {
		return nil
	} else {
		return append(cycle, destination)
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

func (board Board) prepareCycleForResolving(cycle []*Region) {
	for _, region := range cycle {
		region.removeUnit()
		region.order = Order{}
		region.partOfCycle = true
	}
}
