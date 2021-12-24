package game

// Delegates resolving of combat to other functions depending on the state of the area.
func (area *BoardArea) resolveCombat() {
	if area.Control == Uncontrolled && !area.Sea {
		// If area is an empty, uncontrolled land area with a single attacker,
		// then the attacker fights the area.
		if area.IsEmpty() && len(area.IncomingMoves) == 1 {
			area.resolveCombatPvE()
			return
		}

		// If uncontrolled area is not empty or has several attackers,
		// then involved units must first fight each other.
		winner, tie := area.resolveCombatPvP(true)

		// Ties are handled by resolveCombatPvP.
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
		area.resolveCombatPvE()
		return
	}

	// If area is conquered, empty and has only one attacker, it automatically succeeds.
	if area.IsEmpty() && len(area.IncomingMoves) == 1 {
		area.IncomingMoves[0].moveAndSucceed()
		return
	}

	// If attacked area has defender or multiple attackers, they must fight.
	winner, tie := area.resolveCombatPvP(true)
	if tie {
		return
	}
	area.resolveWinner(winner)
}

// Resolves combat between a single attacker and an unconquered area.
func (area *BoardArea) resolveCombatPvE() {
	// Assumes check has already been made that there is just one attacker.
	order := area.IncomingMoves[0]

	win := area.calculateCombatPvE(order)

	if win {
		order.moveAndSucceed()
	} else {
		order.failMove()
	}
}

// Resolves PvE combat in an area if it results in a loss, but leaves wins unresolved.
// Takes in the order with which to calculate combat, and returns whether the order was resolved.
func (area *BoardArea) resolveCombatPvELoss(order *Order) (resolved bool) {
	if area.Control != Uncontrolled {
		return false
	}

	win := area.calculateCombatPvE(order)

	if !win {
		order.failMove()
		resolved = true
	}

	return resolved
}

// Takes in an order (assuming it's a move order to the given area)
// and returns whether the move wins in combat against the uncontrolled area.
func (area *BoardArea) calculateCombatPvE(order *Order) bool {
	mods := map[Player][]Modifier{
		order.Player: attackModifiers(*order, false, false, true),
	}

	appendSupportMods(mods, *area, area.IncomingMoves, true)

	combat, result, _ := combatResults(mods)
	area.Combats = append(area.Combats, combat)

	return result.Total >= 4
}

// Resolves combat when attacked area is defended or has multiple attackers.
// Takes in parameter for whether to account for defender in combat (most often true).
// Returns winner ("" in the case of tie) and whether there was a tie for the highest result.
func (area *BoardArea) resolveCombatPvP(includeDefender bool) (Player, bool) {
	mods := make(map[Player][]Modifier)

	for _, move := range area.IncomingMoves {
		mods[move.Player] = attackModifiers(*move, true, false, includeDefender)
	}

	if !area.IsEmpty() && includeDefender {
		mods[area.Unit.Player] = defenseModifiers(*area)
	}

	appendSupportMods(mods, *area, area.IncomingMoves, includeDefender)

	combat, winner, tie := combatResults(mods)
	area.Combats = append(area.Combats, combat)

	// In the case of tie, all moves fail. If more than 2 combatants are involved,
	// all combatants with a result lower than the tie die.
	if tie {
		for _, order := range area.IncomingMoves {
			order.failMove()

			for _, result := range combat {
				if order.Player == result.Player && result.Total < winner.Total {
					order.killAttacker()
				}
			}
		}

		if !area.IsEmpty() && includeDefender {
			for _, result := range combat {
				if area.Unit.Player == result.Player && result.Total < winner.Total {
					area.removeUnit()
				}
			}
		}

		return "", tie
	}

	return winner.Player, tie
}

// Resolves combat when units from two areas attack each other simultaneously.
func resolveBorderCombat(area1 *BoardArea, area2 *BoardArea) {
	mods := make(map[Player][]Modifier)

	for _, area := range []*BoardArea{area1, area2} {
		mods[area.Unit.Player] = attackModifiers(*area.Order, true, true, false)

		appendSupportMods(mods, *area.Order.To, []*Order{area.Order}, false)
	}

	combat, winner, tie := combatResults(mods)
	area1.Combats = append(area1.Combats, combat)
	area2.Combats = append(area2.Combats, combat)

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
