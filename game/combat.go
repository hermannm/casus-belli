package game

import (
	. "immerse-ntnu/hermannia/server/types"
)

func resolveCombat(area *BoardArea) bool {
	if area.Control == Uncontrolled {
		if len(area.IncomingMoves) == 1 {
			resolveCombatPvE(area)
		} else {
			resolved := resolveCombatPvP(area)
			if !resolved {
				return resolved
			}
			resolveCombatPvE(area)
		}
	}

	return true
}

func resolveCombatPvP(area *BoardArea) bool {
	defense := area.Unit != nil
	return defense
}

func resolveCombatPvE(area *BoardArea) {
	order := area.IncomingMoves[0]

	mods := AttackModifiers(*order, false)

	mods = append(mods, Modifier{
		Type:  DiceMod,
		Value: RollDice(),
	})

	for _, supportOrder := range area.IncomingSupports {
		if callSupport(supportOrder, area) == order.Player.Color {
			mods = append(mods, Modifier{
				Type:        SupportMod,
				Value:       1,
				SupportFrom: supportOrder.Player.Color,
			})
		}
	}

	total := 0
	for _, mod := range mods {
		total += mod.Value
	}

	area.Combats = append(area.Combats, Combat{
		Result{
			Total:  total,
			Parts:  mods,
			Player: order.From.Unit.Color,
		},
	})

	if total <= 4 {
		succeedMove(area, order)
	} else {
		failMove(order)
	}
}

func callSupport(supportOrder *Order, area *BoardArea) PlayerColor {
	if supportOrder.Player.Color == area.Control {
		return supportOrder.Player.Color
	}

	for _, move := range area.IncomingMoves {
		if supportOrder.Player.Color == move.From.Unit.Color {
			return supportOrder.Player.Color
		}
	}

	// TODO: implement support dispatch
	return supportOrder.Player.Color
}
