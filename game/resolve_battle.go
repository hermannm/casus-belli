package game

func (board Board) resolveBattle(battle Battle) {
	if battle.IsBorderConflict() {
		board.resolveBorderConflict(battle)
		return
	}

	if len(battle.Results) == 1 {
		board.resolvePvEBattle(battle)
		return
	}

	board.resolvePvPBattle(battle)
}

func (board Board) resolveBorderConflict(battle Battle) {
	winners, _ := battle.parseResults()
	move1 := battle.Results[0].Move
	move2 := battle.Results[1].Move

	if len(winners) != 1 {
		board.removeMove(move1)
		board.removeMove(move2)
		return
	}

	winner := winners[0]

	for _, move := range []Order{move1, move2} {
		if move.Player == winner {
			if board[move.To].IsControlled() {
				board.succeedMove(move)
			}
		} else {
			board.removeMove(move)
			board.removeOriginUnit(move)
		}
	}
}

func (board Board) resolvePvEBattle(battle Battle) {
	winners, _ := battle.parseResults()
	move := battle.Results[0].Move

	if len(winners) != 1 {
		board.removeMove(move)
	}

	board.succeedMove(move)
}

func (board Board) resolvePvPBattle(battle Battle) {
	winners, losers := battle.parseResults()
	tie := len(winners) != 1

	for _, result := range battle.Results {
		move := result.Move
		lost := containsPlayer(losers, move.Player)

		if !lost && !tie {
			if board[move.To].IsControlled() {
				board.succeedMove(move)
			}
		} else {
			board.removeMove(move)

			if lost {
				board.removeOriginUnit(move)
			}
		}
	}
}

// Parses the results of the battle and finds the winners and losers.
// In case of a battle against an unconquered area or a danger zone,
// only one player is returned in one of the lists.
// In case of a battle between players, multiple winners are returned
// in the case of a tie.
func (battle Battle) parseResults() (winners []Player, losers []Player) {
	// Checks if the battle was against an unconquered area.
	if len(battle.Results) == 1 {
		result := battle.Results[0]

		// Order successfully crosses danger zone if it rolls higher than 2.
		if battle.DangerZone != "" {
			if result.Total > 2 {
				return []Player{result.Move.Player}, make([]Player, 0)
			} else {
				return make([]Player, 0), []Player{result.Move.Player}
			}
		}

		// Order conquers uncontrolled area on a result of 4 or higher.
		if result.Total >= 4 {
			return []Player{result.Move.Player}, make([]Player, 0)
		} else {
			return make([]Player, 0), []Player{result.Move.Player}
		}
	}

	// Finds the highest result among the players in the battle.
	highestResult := 0
	for _, result := range battle.Results {
		if result.Total > highestResult {
			highestResult = result.Total
		}
	}

	// Sorts combatants based on whether they exceeded the highest result.
	winners = make([]Player, 0)
	losers = make([]Player, 0)
	for _, result := range battle.Results {
		if result.Total >= highestResult {
			winners = append(winners, result.Move.Player)
		} else {
			losers = append(losers, result.Move.Player)
		}
	}

	return winners, losers
}
