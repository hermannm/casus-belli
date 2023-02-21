package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

// Moves the unit of the given move order to its destination, killing any unit that may have already
// been there, and sets control of the region to the order's player.
//
// Then removes references to this move on the board, and removes any potential order from the
// destination region.
func succeedMove(move gametypes.Order, board gametypes.Board) {
	destination := board.Regions[move.Destination]

	destination.Unit = move.Unit
	destination.Order = gametypes.Order{}
	if !destination.Sea {
		destination.ControllingPlayer = move.Player
	}

	board.Regions[move.Destination] = destination

	board.RemoveUnit(move.Unit, move.Origin)
	board.RemoveOrder(move)
}

// Attempts to move the unit of the given move order back to its origin.
// Returns whether the retreat succeeded.
func attemptRetreat(move gametypes.Order, board gametypes.Board) bool {
	origin := board.Regions[move.Origin]

	if origin.Unit == move.Unit {
		return true
	}

	if origin.IsAttacked() {
		return false
	}

	origin.Unit = move.Unit
	board.Regions[move.Origin] = origin
	return true
}
