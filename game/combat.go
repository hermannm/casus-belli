package game

func (area *BoardArea) resolveCombat() bool {
	if area.Control == Uncontrolled {
		if len(area.IncomingMoves) == 1 {
			area.resolveCombatPvE()
		} else {
			resolved := area.resolveCombatPvP()
			if !resolved {
				return resolved
			}
			area.resolveCombatPvE()
		}
	}

	return true
}

func (area *BoardArea) resolveCombatPvE() {
	order := getOnlyOrder(area.IncomingMoves)

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

func (area *BoardArea) resolveCombatPvP() bool {
	defense := area.Unit != nil
	return defense
}

func resolveBorderCombat(area1 *BoardArea, area2 *BoardArea) {
	mods := make(map[PlayerColor][]Modifier)

	for _, area := range []*BoardArea{area1, area2} {
		mods[area.Unit.Color] = AttackModifiers(*area.Outgoing, true, true)

		appendSupportMods(mods, area, map[string]*Order{area.Name: area.Outgoing})
	}

	combat, winner, tie := combatResults(mods)
	area1.Combats = append(area1.Combats, combat)
	area2.Combats = append(area2.Combats, combat)

	if tie {
		area1.Outgoing.failMove()
		area2.Outgoing.failMove()
	} else {
		if winner.Player == area1.Unit.Color {
			area2.Outgoing.failMove()
			area1.Outgoing.succeedMove()
		} else {
			area1.Outgoing.failMove()
			area2.Outgoing.succeedMove()
		}
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

func appendSupportMods(mods map[PlayerColor][]Modifier, area *BoardArea, moves map[string]*Order) {
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

func callSupport(support *Order, area *BoardArea, moves map[string]*Order) PlayerColor {
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
