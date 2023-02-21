package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

// Resolves effects of the given battle on the board.
// Forwards the given battle to the appropriate battle resolver based on its type.
// Returns any retreating move orders that could not be resolved.
func resolveBattle(battle gametypes.Battle, board gametypes.Board) (retreats []gametypes.Order) {
	if battle.IsBorderConflict() {
		return resolveBorderBattle(battle, board)
	}

	if len(battle.Results) == 1 {
		return resolveSingleplayerBattle(battle, board)
	}

	return resolveMultiplayerBattle(battle, board)
}

// Resolves effects on the board from the given border battle.
// Assumes that the battle consists of exactly 2 results, for each of the regions in the battle,
// that each result is tied to a move order, and that the battle had at least one winner.
// Returns any retreating move orders that could not be resolved.
func resolveBorderBattle(
	battle gametypes.Battle, board gametypes.Board,
) (retreats []gametypes.Order) {
	winners, _ := battle.WinnersAndLosers()
	move1 := battle.Results[0].Move
	move2 := battle.Results[1].Move

	// If there is more than one winner, the battle was a tie, and both moves retreat.
	if len(winners) > 1 {
		board.RemoveOrder(move1)
		board.RemoveOrder(move2)

		if !attemptRetreat(move1, board) {
			retreats = append(retreats, move1)
		}
		if !attemptRetreat(move2, board) {
			retreats = append(retreats, move2)
		}

		return retreats
	}

	winner := winners[0]

	for _, move := range []gametypes.Order{move1, move2} {
		if move.Player == winner {
			// If destination region is uncontrolled, the player must win a singleplayer battle
			// there before taking control.
			if board.Regions[move.Destination].IsControlled() {
				succeedMove(move, board)
			}
		} else {
			board.RemoveOrder(move)
			board.RemoveUnit(move.Unit, move.Origin)
		}
	}

	return nil
}

// Resolves effects on the board from the given singleplayer battle (player vs. neutral region).
// Assumes that the battle has a single result, with a move order tied to it.
// Returns the move order in a list if it fails retreat, or nil otherwise.
func resolveSingleplayerBattle(
	battle gametypes.Battle, board gametypes.Board,
) (retreats []gametypes.Order) {
	winners, _ := battle.WinnersAndLosers()
	move := battle.Results[0].Move

	if len(winners) != 1 {
		board.RemoveOrder(move)

		if attemptRetreat(move, board) {
			return nil
		} else {
			return []gametypes.Order{move}
		}
	}

	succeedMove(move, board)
	return nil
}

// Resolves effects on the board from the given multiplayer battle.
// Assumes that the battle has at least 1 winner.
// Returns any retreating move orders that could not be resolved.
func resolveMultiplayerBattle(
	battle gametypes.Battle, board gametypes.Board,
) (retreats []gametypes.Order) {
	winners, losers := battle.WinnersAndLosers()
	tie := len(winners) != 1

	for _, result := range battle.Results {
		// If the result has a DefenderRegion, it is the result of the region's defender.
		// If the defender won, nothing changes for them.
		// If an attacker won, changes to the defender will be handled by calling succeedMove.
		if result.DefenderRegion != "" {
			continue
		}

		move := result.Move

		lost := false
		for _, otherPlayer := range losers {
			if otherPlayer == move.Player {
				lost = true
			}
		}

		if lost {
			board.RemoveOrder(move)
			board.RemoveUnit(move.Unit, move.Origin)
			continue
		}

		if tie {
			board.RemoveOrder(move)
			if !attemptRetreat(move, board) {
				retreats = append(retreats, move)
			}
			continue
		}

		if board.Regions[move.Destination].IsControlled() {
			succeedMove(move, board)
		}
	}

	return retreats
}
