package gameboard

// Resolves moves to the given region on the board.
// Assumes that the region has incoming moves (moveCount > 0).
//
// Immediately resolves regions that do not require battle, and adds them to the given processed
// map.
//
// Adds embattled regions to the given processing map, and forwards them to appropriate battle
// calculation functions, which send results to the given battleReceiver.
//
// Skips regions that have outgoing moves, unless they are part of a move cycle.
// If allowPlayerConflict is false, skips regions that require battle between players.
func (board Board) resolveRegionMoves(
	region Region,
	moveCount int,
	allowPlayerConflict bool,
	battleReceiver chan Battle,
	processing map[string]struct{},
	processed map[string]struct{},
	messenger Messenger,
) {
	// Finds out if the move is part of a two-way cycle (moves moving against each other),
	// and resolves it.
	twoWayCycle, region2, samePlayer := board.discoverTwoWayCycle(region)
	if twoWayCycle {
		if samePlayer {
			// If both moves are by the same player, removes the units from their origin regions,
			// as they may not be allowed to retreat if their origin region is taken.
			for _, cycleRegion := range [2]Region{region, region2} {
				cycleRegion.Unit = Unit{}
				cycleRegion.Order = Order{}
				board.Regions[cycleRegion.Name] = cycleRegion
			}
		} else {
			// If the moves are from different players, they battle in the middle.
			go calculateBorderBattle(region, region2, battleReceiver, messenger)
			processing[region.Name], processing[region2.Name] = struct{}{}, struct{}{}
			return
		}
	} else {
		// If there is a cycle longer than 2 moves, forwards the resolving to 'resolveCycle'.
		cycle, _ := board.discoverCycle(region.Order, region.Name)
		if cycle != nil {
			board.resolveCycle(
				cycle,
				allowPlayerConflict,
				battleReceiver,
				processing,
				processed,
				messenger,
			)
			return
		}
	}

	// Empty regions with only a single incoming move are either auto-successes or a singleplayer
	// battle.
	if moveCount == 1 && region.IsEmpty() {
		move := region.IncomingMoves[0]

		if region.IsControlled() || region.Sea {
			board.succeedMove(move)
			processed[region.Name] = struct{}{}
			return
		}

		go region.calculateSingleplayerBattle(move, battleReceiver, messenger)
		processing[region.Name] = struct{}{}
		return
	}

	// If the destination region has an outgoing move order, that must be resolved first.
	if region.Order.Type == OrderMove {
		return
	}

	// If the function has not returned yet, then it must be a multiplayer battle.
	go region.calculateMultiplayerBattle(!region.IsEmpty(), battleReceiver, messenger)
	processing[region.Name] = struct{}{}
}

// Calculates battle between a single attacker and an unconquered region.
// Sends the resulting battle to the given battleReceiver.
func (region Region) calculateSingleplayerBattle(
	move Order,
	battleReceiver chan<- Battle,
	messenger Messenger,
) {
	playerResults := map[string]Result{
		move.Player: {Parts: move.attackModifiers(region, false, false, true), Move: move},
	}

	appendSupportMods(playerResults, region, false, messenger)

	battleReceiver <- Battle{Results: calculateTotals(playerResults)}
}

// Calculates battle when attacked region is defended or has multiple attackers.
// Takes in parameter for whether to account for defender in battle (most often true).
// Sends the resulting battle to the given battleReceiver.
func (region Region) calculateMultiplayerBattle(
	includeDefender bool,
	battleReceiver chan<- Battle,
	messenger Messenger,
) {
	playerResults := make(map[string]Result)

	for _, move := range region.IncomingMoves {
		playerResults[move.Player] = Result{
			Parts: move.attackModifiers(region, true, false, includeDefender),
			Move:  move,
		}
	}

	if !region.IsEmpty() && includeDefender {
		playerResults[region.Unit.Player] = Result{
			Parts:          region.defenseModifiers(),
			DefenderRegion: region.Name,
		}
	}

	appendSupportMods(playerResults, region, includeDefender, messenger)

	battleReceiver <- Battle{Results: calculateTotals(playerResults)}
}

// Calculates battle when units from two regions attack each other simultaneously.
// Sends the resulting battle to the given battleReceiver.
func calculateBorderBattle(region1 Region, region2 Region, battleReceiver chan<- Battle, messenger Messenger) {
	move1 := region1.Order
	move2 := region2.Order
	playerResults := map[string]Result{
		move1.Player: {Parts: move1.attackModifiers(region2, true, true, false), Move: move1},
		move2.Player: {Parts: move2.attackModifiers(region1, true, true, false), Move: move2},
	}

	appendSupportMods(playerResults, region2, false, messenger)
	appendSupportMods(playerResults, region1, false, messenger)

	battleReceiver <- Battle{Results: calculateTotals(playerResults)}
}
