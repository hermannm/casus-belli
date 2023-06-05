package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

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
// Returns whether the region is a part of a cycle, and if so, the second region in the cycle,
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