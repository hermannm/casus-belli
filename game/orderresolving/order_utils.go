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
	to := board.Regions[move.To]
	to.Unit = move.Unit
	to.Order = gametypes.Order{}
	if !to.Sea {
		to.ControllingPlayer = move.Player
	}
	board.Regions[move.To] = to

	board.RemoveUnit(move.Unit, move.From)

	board.RemoveOrder(move)
}

// Attempts to move the unit of the given move order back to its origin.
// Returns whether the retreat succeeded.
func attemptRetreat(move gametypes.Order, board gametypes.Board) bool {
	from := board.Regions[move.From]

	if from.Unit == move.Unit {
		return true
	}

	if len(from.IncomingMoves) != 0 {
		return false
	}

	from.Unit = move.Unit
	board.Regions[move.From] = from
	return true
}
