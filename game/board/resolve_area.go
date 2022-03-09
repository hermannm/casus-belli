package board

// Resolves moves to the given area on the board.
// Assumes that the area has incoming moves (moveCount > 0).
//
// Immediately resolves areas that do not require battle,
// and adds them to the given processed map.
//
// Adds embattled areas to the given processing map,
// and forwards them to appropriate battle calculators,
// which send results to the given battleReceiver.
//
// Skips areas that have outgoing moves, unless they are part of a move cycle.
// If playerConflictsAllowed is true, skips areas that require battle between players.
func (board Board) resolveAreaMoves(
	area Area,
	moveCount int,
	playerConflictsAllowed bool,
	battleReceiver chan Battle,
	processing map[string]struct{},
	processed map[string]struct{},
) {
	// Finds out if the move is part of a two-way cycle
	// (moves moving against each other), and resolves it.
	twoWayCycle, area2, samePlayer := board.discoverTwoWayCycle(area)
	if twoWayCycle {
		if samePlayer {
			// If both moves are by the same player, removes the units from their origin areas,
			// as they may not be allowed to retreat if their origin area is taken.
			for _, cycleArea := range [2]Area{area, area2} {
				cycleArea = cycleArea.setUnit(Unit{})
				cycleArea = cycleArea.setOrder(Order{})
				board[cycleArea.Name] = cycleArea
			}
		} else {
			// If the moves are from different players, they battle in the middle.
			go calculateBorderBattle(area, area2, battleReceiver)
			processing[area.Name], processing[area2.Name] = struct{}{}, struct{}{}
			return
		}
	} else {
		// If there is a cycle longer than 2 moves, forwards the resolving to 'resolveCycle'.
		cycle, _ := board.discoverCycle(area.Order, area.Name)
		if cycle != nil {
			board.resolveCycle(cycle, playerConflictsAllowed, battleReceiver, processing, processed)
			return
		}
	}

	// Empty areas with only a single incoming move are either auto-successes or a singleplayer battle.
	if moveCount == 1 && area.IsEmpty() {
		move := area.IncomingMoves[0]

		if area.IsControlled() || area.Sea {
			board.succeedMove(move)
			processed[area.Name] = struct{}{}
			return
		}

		go area.calculateSingleplayerBattle(move, battleReceiver)
		processing[area.Name] = struct{}{}
		return
	}

	// If the destination area has an outgoing move order,
	// that must be resolved first.
	if area.Order.Type == OrderMove {
		return
	}

	// If the function has not returned yet, then it must be a multiplayer battle.
	go area.calculateMultiplayerBattle(!area.IsEmpty(), battleReceiver)
	processing[area.Name] = struct{}{}
}

// Calculates battle between a single attacker and an unconquered area.
// Sends the resulting battle to the given battleReceiver.
func (area Area) calculateSingleplayerBattle(move Order, battleReceiver chan<- Battle) {
	results := map[Player]Result{
		move.Player: {
			Parts: move.attackModifiers(area, false, false, true),
			Move:  move,
		},
	}

	appendSupportMods(results, area, false)

	battleReceiver <- Battle{
		Results: calculateTotals(results),
	}
}

// Calculates battle when attacked area is defended or has multiple attackers.
// Takes in parameter for whether to account for defender in battle (most often true).
// Sends the resulting battle to the given battleReceiver.
func (area Area) calculateMultiplayerBattle(includeDefender bool, battleReceiver chan<- Battle) {
	results := make(map[Player]Result)

	for _, move := range area.IncomingMoves {
		results[move.Player] = Result{
			Parts: move.attackModifiers(area, true, false, includeDefender),
			Move:  move,
		}
	}

	if !area.IsEmpty() && includeDefender {
		results[area.Unit.Player] = Result{
			Parts:        area.defenseModifiers(),
			DefenderArea: area.Name,
		}
	}

	appendSupportMods(results, area, includeDefender)

	battleReceiver <- Battle{
		Results: calculateTotals(results),
	}
}

// Calculates battle when units from two areas attack each other simultaneously.
// Sends the resulting battle to the given battleReceiver.
func calculateBorderBattle(area1 Area, area2 Area, battleReceiver chan<- Battle) {
	move1 := area1.Order
	move2 := area2.Order
	results := map[Player]Result{
		move1.Player: {
			Parts: move1.attackModifiers(area2, true, true, false),
			Move:  move1,
		},
		move2.Player: {
			Parts: move2.attackModifiers(area1, true, true, false),
			Move:  move2,
		},
	}

	appendSupportMods(results, area2, false)
	appendSupportMods(results, area1, false)

	battleReceiver <- Battle{
		Results: calculateTotals(results),
	}
}
