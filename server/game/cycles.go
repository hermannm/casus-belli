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

func (board Board) prepareCycleForResolving(cycle []*Region) {
	for _, region := range cycle {
		region.removeUnit()
		region.order = Order{}
		region.partOfCycle = true
	}
}
