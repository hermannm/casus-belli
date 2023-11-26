package game

// Recursively finds a cycle of move orders through regions starting and ending with the given
// firstRegionName.
func (board Board) discoverCycle(firstRegionName RegionName, region *Region) (cycle []*Region) {
	if region.order.Type != OrderMove {
		return nil
	}

	// The base case: the destination is the beginning of the cycle.
	if region.order.Destination == firstRegionName {
		return []*Region{region}
	}

	// If the base case is not yet reached, passes cycle discovery to the next region in the chain.
	cycle = board.discoverCycle(firstRegionName, board[region.order.Destination])
	if cycle == nil {
		return nil
	} else {
		return append(cycle, region)
	}
}

func (board Board) prepareCycleForResolving(cycle []*Region) {
	for _, region := range cycle {
		region.removeUnit()
		region.order = Order{}
		region.partOfCycle = true
	}
}
