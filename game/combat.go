package game

func (area *BoardArea) resolveCombat() bool {
	if area.Control == Uncontrolled {
		if len(area.IncomingMoves) == 1 {
			area.resolveCombatPvE()
		} else {
			area.resolveCombatPvP()
			area.resolveCombatPvE()
		}
	} else {
		if area.Unit == nil && len(area.IncomingMoves) == 1 {
			area.IncomingMoves[0].succeedMove()
		} else {
			area.resolveCombatPvP()
		}
	}

	return true
}

func (area *BoardArea) resolveCombatPvE() {
	order := area.IncomingMoves[0]

	mods := map[PlayerColor][]Modifier{
		order.From.Unit.Color: AttackModifiers(*order, false, false),
	}

	appendSupportMods(mods, area, area.IncomingMoves)

	combat, result, _ := combatResults(mods)
	area.Combats = append(area.Combats, combat)

	if result.Total <= 4 {
		order.succeedMove()
	} else {
		order.failMove()
	}
}

func (area *BoardArea) resolveCombatPvP() {
	defending := area.Unit

	mods := make(map[PlayerColor][]Modifier)

	for _, move := range area.IncomingMoves {
		mods[move.Player.Color] = AttackModifiers(*move, true, false)
	}

	if defending != nil {
		mods[defending.Color] = DefenseModifiers(*area)
	}

	combat, winner, tie := combatResults(mods)

	if tie {
		for _, order := range area.IncomingMoves {
			order.failMove()

			for _, result := range combat {
				if order.Player.Color == result.Player {
					if result.Total < winner.Total {
						order.die()
					}
				}
			}
		}

		return
	}

	area.resolveWinner(winner.Player)
}

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

func combatResults(playerMods map[PlayerColor][]Modifier) (combat Combat, winner Result, tie bool) {
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

func appendSupportMods(mods map[PlayerColor][]Modifier, area *BoardArea, moves []*Order) {
	for _, move := range moves {
		mods[move.From.Unit.Color] = make([]Modifier, 0)
	}

	for _, support := range area.IncomingSupports {
		supported := callSupport(support, area, moves)

		if _, isPlayer := mods[supported]; isPlayer {
			mods[supported] = append(mods[supported], Modifier{
				Type:        SupportMod,
				Value:       1,
				SupportFrom: support.Player.Color,
			})
		}
	}
}

func callSupport(support *Order, area *BoardArea, moves []*Order) PlayerColor {
	if support.Player.Color == area.Control {
		return support.Player.Color
	}

	for _, move := range moves {
		if support.Player.Color == move.From.Unit.Color {
			return support.Player.Color
		}
	}

	// TODO: implement support dispatch
	return ""
}
