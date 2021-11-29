package game

// Delegates resolving of combat to other functions depending on the state of the area.
func (area *BoardArea) resolveCombat() {
	if area.Control == Uncontrolled && !area.Sea {
		if area.Unit == nil && len(area.IncomingMoves) == 1 {
			// If area is an empty, uncontrolled land area with a single attacker,
			// then the attacker fights the area.
			area.resolveCombatPvE()
		} else {
			// If uncontrolled area is not empty or has several attackers,
			// then involved units must first fight each other.
			winner, tie := area.resolveCombatPvP()

			// Consequences of ties are handled by resolveCombatPvP.
			if !tie {
				// If area was already occupied and occupier won, it stays there.
				// If an attacker won, they get to attempt to conquer the area.
				if area.Unit != nil && area.Unit.Color == winner {
					area.resolveWinner(winner)
				} else {
					area.resolveIntermediaryWinner(winner)
					area.resolveCombatPvE()
				}
			}
		}
	} else {
		// If area is conquered, empty and has only one attacker, it automatically succeeds.
		if area.Unit == nil && len(area.IncomingMoves) == 1 {
			area.IncomingMoves[0].succeedMove()
		} else {
			winner, tie := area.resolveCombatPvP()

			if !tie {
				area.resolveWinner(winner)
			}
		}
	}
}

// Resolves combat between a single attacker and an unconquered area.
func (area *BoardArea) resolveCombatPvE() {
	// Assumes check has already been made that there is just one attacker.
	order := area.IncomingMoves[0]

	mods := map[PlayerColor][]Modifier{
		order.From.Unit.Color: AttackModifiers(*order, false, false),
	}

	appendSupportMods(mods, area, area.IncomingMoves)

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
func (area *BoardArea) resolveCombatPvP() (PlayerColor, bool) {
	defending := area.Unit

	mods := make(map[PlayerColor][]Modifier)

	for _, move := range area.IncomingMoves {
		mods[move.Player] = AttackModifiers(*move, true, false)
	}

	if defending != nil {
		mods[defending.Color] = DefenseModifiers(*area)
	}

	appendSupportMods(mods, area, area.IncomingMoves)

	combat, winner, tie := combatResults(mods)
	area.Combats = append(area.Combats, combat)

	// In the case of tie, all moves fail. If more than 2 combatants are involved,
	// all combatants with a result lower than the tie die.
	if tie {
		for _, order := range area.IncomingMoves {
			order.failMove()

			for _, result := range combat {
				if order.Player == result.Player {
					if result.Total < winner.Total {
						order.die()
					}
				}
			}
		}

		if defending != nil {
			for _, result := range combat {
				if defending.Color == result.Player {
					if result.Total < winner.Total {
						area.killDefender()
					}
				}
			}
		}

		return "", tie
	}

	return winner.Player, tie
}

// Resolves combat when units from two areas attack each other simultaneously.
func resolveBorderCombat(area1 *BoardArea, area2 *BoardArea) {
	mods := make(map[PlayerColor][]Modifier)

	for _, area := range []*BoardArea{area1, area2} {
		mods[area.Unit.Color] = AttackModifiers(*area.Outgoing, true, true)

		appendSupportMods(mods, area.Outgoing.To, []*Order{area.Outgoing})
	}

	combat, winner, tie := combatResults(mods)
	area1.Combats = append(area1.Combats, combat)
	area2.Combats = append(area2.Combats, combat)

	if tie {
		area1.Outgoing.failMove()
		area2.Outgoing.failMove()
		return
	}

	if winner.Player == area1.Unit.Color {
		area2.Outgoing.failMove()
		area1.Outgoing.succeedMove()
	} else {
		area1.Outgoing.failMove()
		area2.Outgoing.succeedMove()
	}
}

// Constructs combat results from combatants' modifiers.
func combatResults(playerMods map[PlayerColor][]Modifier) (
	combat Combat,
	winner Result,
	tie bool,
) {
	for player, mods := range playerMods {
		total := modTotal(mods)

		result := Result{
			Total:  modTotal(mods),
			Parts:  mods,
			Player: player,
		}

		if total > winner.Total {
			winner = result
			tie = false
		} else if total == winner.Total {
			tie = true
		}

		combat = append(combat, result)
	}

	return combat, winner, tie
}

// Calls support for a combat and adds support modifiers appropriately.
func appendSupportMods(mods map[PlayerColor][]Modifier, area *BoardArea, moves []*Order) {
	for _, support := range area.IncomingSupports {
		supported := callSupport(support, area, moves)

		if _, isPlayer := mods[supported]; isPlayer {
			mods[supported] = append(mods[supported], Modifier{
				Type:        SupportMod,
				Value:       1,
				SupportFrom: support.Player,
			})
		}
	}
}

// Returns which player a given support order supports in a combat.
// If combatant matches support order's player, support is automatically given.
// TODO: Implement asking player who to support if they are not involved themselves.
func callSupport(support *Order, area *BoardArea, moves []*Order) PlayerColor {
	if area.Unit != nil && area.Unit.Color == support.Player {
		return support.Player
	}

	for _, move := range moves {
		if support.Player == move.From.Unit.Color {
			return support.Player
		}
	}

	return ""
}
