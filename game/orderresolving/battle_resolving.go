package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

// Resolves effects of the given battle on the board.
// Forwards the given battle to the appropriate battle resolver based on its type.
// Returns any retreating move orders that could not be resolved.
func (resolver *MoveResolver) resolveBattle(battle gametypes.Battle, board gametypes.Board) {
	if battle.IsBorderConflict() {
		resolver.resolveBorderBattle(battle, board)
	} else if len(battle.Results) == 1 {
		resolver.resolveSingleplayerBattle(battle, board)
	} else {
		resolver.resolveMultiplayerBattle(battle, board)
	}

	resolver.resolvedBattles = append(resolver.resolvedBattles, battle)

	for _, region := range battle.RegionNames() {
		resolver.resolvingRegions.Remove(region)
	}
}

// Resolves effects on the board from the given border battle.
// Assumes that the battle consists of exactly 2 results, for each of the regions in the battle,
// that each result is tied to a move order, and that the battle had at least one winner.
// Returns any retreating move orders that could not be resolved.
func (resolver *MoveResolver) resolveBorderBattle(battle gametypes.Battle, board gametypes.Board) {
	winners, losers := battle.WinnersAndLosers()
	move1 := battle.Results[0].Move
	move2 := battle.Results[1].Move

	// If there is more than one winner, the battle was a tie, and both moves retreat.
	if len(winners) > 1 {
		board.RemoveOrder(move1)
		board.RemoveOrder(move2)

		resolver.attemptRetreat(move1, board)
		resolver.attemptRetreat(move2, board)

		return
	}

	loser := losers[0]

	for _, move := range []gametypes.Order{move1, move2} {
		// Only the loser is affected by the results of the border battle; the winner may still have
		// to win a battle in the destination region, which will be handled by the next cycle of the
		// move resolver.
		if move.Player == loser {
			board.RemoveOrder(move)
			board.RemoveUnit(move.Unit, move.Origin)
		}
	}
}

// Resolves effects on the board from the given singleplayer battle (player vs. neutral region).
// Assumes that the battle has a single result, with a move order tied to it.
func (resolver *MoveResolver) resolveSingleplayerBattle(
	battle gametypes.Battle, board gametypes.Board,
) {
	winners, _ := battle.WinnersAndLosers()
	move := battle.Results[0].Move

	if len(winners) != 1 {
		board.RemoveOrder(move)
		resolver.attemptRetreat(move, board)
		return
	}

	resolver.succeedMove(move, board)
}

// Resolves effects on the board from the given multiplayer battle.
// Assumes that the battle has at least 1 winner.
func (resolver *MoveResolver) resolveMultiplayerBattle(
	battle gametypes.Battle, board gametypes.Board,
) {
	winners, losers := battle.WinnersAndLosers()
	tie := len(winners) != 1

	for _, result := range battle.Results {
		// If the result has a DefenderRegion, it is the result of the region's defender.
		// If the defender won or tied, nothing changes for them.
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
			resolver.attemptRetreat(move, board)
			continue
		}

		if board.Regions[move.Destination].IsControlled() {
			resolver.succeedMove(move, board)
		}
	}
}

// Attempts to move the unit of the given move order back to its origin.
// Returns whether the retreat succeeded.
func (resolver *MoveResolver) attemptRetreat(move gametypes.Order, board gametypes.Board) {
	origin := board.Regions[move.Origin]

	if origin.Unit == move.Unit {
		return
	}

	if origin.IsAttacked() {
		resolver.retreats[move.Origin] = move
		return
	}

	origin.Unit = move.Unit
	board.Regions[move.Origin] = origin
}
