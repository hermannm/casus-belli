package game

// Resolves effects of the given battle on the board.
// Forwards the given battle to the appropriate battle resolver based on its type.
// Returns any retreating move orders that could not be resolved.
func (board Board) resolveBattle(battle Battle) (retreats []Order) {
	if battle.isBorderConflict() {
		return board.resolveBorderBattle(battle)
	}

	if len(battle.Results) == 1 {
		return board.resolveSingleplayerBattle(battle)
	}

	return board.resolveMultiplayerBattle(battle)
}

// Resolves effects on the board from the given border battle.
// Assumes that the battle consists of exactly 2 results, for each of the areas in the battle,
// that each result is tied to a move order, and that the battle had at least one winner.
// Returns any retreating move orders that could not be resolved.
func (board Board) resolveBorderBattle(battle Battle) (retreats []Order) {
	winners, _ := battle.parseResults()
	move1 := battle.Results[0].Move
	move2 := battle.Results[1].Move

	// If there is more than one winner, the battle was a tie, and both moves retreat.
	if len(winners) > 1 {
		board.removeMove(move1)
		board.removeMove(move2)

		if !board.attemptRetreat(move1) {
			retreats = append(retreats, move1)
		}
		if !board.attemptRetreat(move2) {
			retreats = append(retreats, move2)
		}

		return retreats
	}

	winner := winners[0]

	for _, move := range []Order{move1, move2} {
		if move.Player == winner {
			// If destination area is uncontrolled, the player must win a singleplayer battle there before taking control.
			if board[move.To].IsControlled() {
				board.succeedMove(move)
			}
		} else {
			board.removeMove(move)
			board.removeOriginUnit(move)
		}
	}

	return nil
}

// Resolves effects on the board from the given singleplayer battle (player vs. neutral area).
// Assumes that the battle has a single result, with a move order tied to it.
// Returns the move order in a list if it fails retreat, or nil otherwise.
func (board Board) resolveSingleplayerBattle(battle Battle) (retreats []Order) {
	winners, _ := battle.parseResults()
	move := battle.Results[0].Move

	if len(winners) != 1 {
		board.removeMove(move)

		if board.attemptRetreat(move) {
			return nil
		} else {
			return []Order{move}
		}
	}

	board.succeedMove(move)
	return nil
}

// Resolves effects on hte board from the given multiplayer battle.
// Assumes that the battle has at least 1 winner
// Returns any retreating move orders that could not be resolved.
func (board Board) resolveMultiplayerBattle(battle Battle) (retreats []Order) {
	winners, losers := battle.parseResults()
	tie := len(winners) != 1

	for _, result := range battle.Results {
		// If the result has a DefenderArea, it is the result of the area's defender.
		// If the defender won, nothing changes for them.
		// If an attacker won, changes to the defender will be handled by calling succeedMove.
		if result.DefenderArea != "" {
			continue
		}

		move := result.Move
		lost := containsPlayer(losers, move.Player)

		if lost {
			board.removeMove(move)
			board.removeOriginUnit(move)
			continue
		}

		if tie {
			board.removeMove(move)
			if !board.attemptRetreat(move) {
				retreats = append(retreats, move)
			}
			continue
		}

		if board[move.To].IsControlled() {
			board.succeedMove(move)
		}
	}

	return retreats
}
