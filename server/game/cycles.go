package game

type MoveCycle []*Region

// Recursively finds a cycle of move orders through regions starting and ending with the given
// firstRegionName.
func (board Board) findCycle(firstRegionName RegionName, region *Region) MoveCycle {
	if region.order.Type != OrderMove {
		return nil
	}

	// The base case: the destination is the beginning of the cycle.
	if region.order.Destination == firstRegionName {
		return []*Region{region}
	}

	// If the base case is not yet reached, passes cycle discovery to the next region in the chain.
	cycle := board.findCycle(firstRegionName, board[region.order.Destination])
	if cycle == nil {
		return nil
	} else {
		return append(cycle, region)
	}
}

func (board Board) findBorderBattle(region *Region) (isBorderBattle bool, secondRegion *Region) {
	if region.order.Type != OrderMove {
		return false, nil
	}

	secondRegion = board[region.order.Destination]
	if secondRegion.order.Type == OrderMove && secondRegion.order.Destination == region.Name &&
		region.order.Faction != secondRegion.order.Faction {
		return true, secondRegion
	}

	return false, nil
}

func (cycle MoveCycle) prepareForResolving() {
	for _, region := range cycle {
		region.removeUnit()
		region.order = Order{}
		region.partOfCycle = true
	}
}
