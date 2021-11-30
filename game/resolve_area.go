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
		winner, tie := area.resolveCombatPvP()

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
		area.IncomingMoves[0].succeedMove()
		return
	}

	// If attacked area has defender or multiple attackers, they must fight.
	winner, tie := area.resolveCombatPvP()
	if tie {
		return
	}
	area.resolveWinner(winner)
}

// Resolves combat between a single attacker and an unconquered area.
func (area *BoardArea) resolveCombatPvE() {
	// Assumes check has already been made that there is just one attacker.
	order := area.IncomingMoves[0]

	mods := map[Player][]Modifier{
		order.From.Unit.Player: attackModifiers(*order, false, false),
	}

	appendSupportMods(mods, *area, area.IncomingMoves)

	combat, result, _ := combatResults(mods)
	area.Combats = append(area.Combats, combat)

	if result.Total >= 4 {
		order.succeedMove()
	} else {
		order.failMove()
	}
}

// Resolves combat when attacked area is defended or has multiple attackers.
// Returns winner ("" in the case of tie) and whether there was a tie for the highest result.
func (area *BoardArea) resolveCombatPvP() (Player, bool) {
	mods := make(map[Player][]Modifier)

	for _, move := range area.IncomingMoves {
		mods[move.Player] = attackModifiers(*move, true, false)
	}

	if !area.IsEmpty() {
		mods[area.Unit.Player] = defenseModifiers(*area)
	}

	appendSupportMods(mods, *area, area.IncomingMoves)

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

		if !area.IsEmpty() {
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
		mods[area.Unit.Player] = attackModifiers(*area.Order, true, true)

		appendSupportMods(mods, *area.Order.To, []*Order{area.Order})
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
		area1.Order.succeedMove()
	} else {
		area1.Order.failMove()
		area2.Order.succeedMove()
	}
}