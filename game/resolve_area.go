package game

// Delegates resolving of battle to other functions depending on the state of the area.
func (area *Area) resolveBattle() {
	if area.Control == Uncontrolled && !area.Sea {
		// If area is an empty, uncontrolled land area with a single attacker,
		// then the attacker fights the area.
		if area.IsEmpty() && len(area.IncomingMoves) == 1 {
			area.resolvePvEBattle()
			return
		}

		// If uncontrolled area is not empty or has several attackers,
		// then involved units must first fight each other.
		winner, tie := area.resolvePvPBattle(true)

		// Ties are handled by resolvePvPBattle.
		if tie {
			return
		}

		// If area was already occupied and occupier won, it stays there.
		if !area.IsEmpty() && area.Unit.Player == winner {
			area.resolveWinner(winner)
			return
		}

		// If an attacker won, they get to attempt to conquer the area.
		area.resolveIntermediaryWinner(winner)
		area.resolvePvEBattle()
		return
	}

	// If area is conquered, empty and has only one attacker, it automatically succeeds.
	if area.IsEmpty() && len(area.IncomingMoves) == 1 {
		area.IncomingMoves[0].moveAndSucceed()
		return
	}

	// If attacked area has defender or multiple attackers, they must fight.
	winner, tie := area.resolvePvPBattle(true)
	if tie {
		return
	}
	area.resolveWinner(winner)
}

// Takes in an order (assuming it's a move order to the given area)
// and returns whether the move wins battle against the uncontrolled area.
func (area *Area) calculatePvEBattle(order *Order) bool {
	mods := map[Player][]Modifier{
		order.Player: attackModifiers(*order, false, false, true),
	}

	appendSupportMods(mods, *area, area.IncomingMoves, true)

	battle, result, _ := battleResults(mods)
	area.Battles = append(area.Battles, battle)

	return result.Total >= 4
}

// Resolves battle between a single attacker and an unconquered area.
func (area *Area) resolvePvEBattle() {
	// Assumes check has already been made that there is just one attacker.
	order := area.IncomingMoves[0]

	win := area.calculatePvEBattle(order)

	if win {
		order.moveAndSucceed()
	} else {
		order.failMove()
	}
}

// Resolves PvE battle in an area if it results in a loss, but leaves wins unresolved.
// Takes in the order with which to calculate battle, and returns whether the order was resolved.
func (area *Area) resolvePvEBattleLoss(order *Order) (resolved bool) {
	if area.Control != Uncontrolled {
		return false
	}

	win := area.calculatePvEBattle(order)

	if !win {
		order.failMove()
		resolved = true
	}

	return resolved
}

// Resolves battle when attacked area is defended or has multiple attackers.
// Takes in parameter for whether to account for defender in battle (most often true).
// Returns winner ("" in the case of tie) and whether there was a tie for the highest result.
func (area *Area) resolvePvPBattle(includeDefender bool) (Player, bool) {
	mods := make(map[Player][]Modifier)

	for _, move := range area.IncomingMoves {
		mods[move.Player] = attackModifiers(*move, true, false, includeDefender)
	}

	if !area.IsEmpty() && includeDefender {
		mods[area.Unit.Player] = defenseModifiers(*area)
	}

	appendSupportMods(mods, *area, area.IncomingMoves, includeDefender)

	battle, winner, tie := battleResults(mods)
	area.Battles = append(area.Battles, battle)

	if !tie {
		return winner.Player, tie
	}

	// In the case of tie, all moves fail. If more than 2 players are involved,
	// all players with a result lower than the tie die.
	for _, order := range area.IncomingMoves {
		order.failMove()

		for _, result := range battle {
			if order.Player == result.Player && result.Total < winner.Total {
				order.killAttacker()
			}
		}
	}

	if !area.IsEmpty() && includeDefender {
		for _, result := range battle {
			if area.Unit.Player == result.Player && result.Total < winner.Total {
				area.removeUnit()
			}
		}
	}

	return "", tie
}

// Resolves battle when units from two areas attack each other simultaneously.
func resolveBorderBattle(area1 *Area, area2 *Area) {
	mods := make(map[Player][]Modifier)

	for _, area := range []*Area{area1, area2} {
		mods[area.Unit.Player] = attackModifiers(*area.Order, true, true, false)

		appendSupportMods(mods, *area.Order.To, []*Order{area.Order}, false)
	}

	battle, winner, tie := battleResults(mods)
	area1.Battles = append(area1.Battles, battle)
	area2.Battles = append(area2.Battles, battle)

	if tie {
		area1.Order.failMove()
		area2.Order.failMove()
		return
	}

	if winner.Player == area1.Unit.Player {
		area2.Order.failMove()
		area1.Order.moveAndSucceed()
	} else {
		area1.Order.failMove()
		area2.Order.moveAndSucceed()
	}
}
