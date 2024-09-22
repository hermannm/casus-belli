package game

type MoveCycle []*Region

// Recursively finds a cycle of move orders through regions starting and ending with the given
// firstRegionName.
func (board Board) findCycle(firstRegionName RegionName, region *Region) MoveCycle {
	order, hasOrder := region.order.Get()
	if !hasOrder || order.Type != OrderMove {
		return nil
	}

	// The base case: the destination is the beginning of the cycle.
	if order.Destination == firstRegionName {
		return []*Region{region}
	}

	// If the base case is not yet reached, passes cycle discovery to the next region in the chain.
	cycle := board.findCycle(firstRegionName, board[order.Destination])
	if cycle == nil {
		return nil
	} else {
		return append(cycle, region)
	}
}

func (board Board) findBorderBattle(region *Region) (isBorderBattle bool, secondRegion *Region) {
	order, hasOrder := region.order.Get()
	if !hasOrder || order.Type != OrderMove {
		return false, nil
	}

	secondRegion = board[order.Destination]
	secondOrder, hasSecondOrder := secondRegion.order.Get()
	if !hasSecondOrder || secondOrder.Type != OrderMove {
		return false, nil
	}

	if secondOrder.Destination == region.Name && order.Faction != secondOrder.Faction {
		return true, secondRegion
	}

	return false, nil
}

func (cycle MoveCycle) prepareForResolving() {
	for _, region := range cycle {
		region.removeUnit()
		region.order.Clear()
		region.partOfCycle = true
	}
}
