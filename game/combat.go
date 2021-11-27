package game

import (
	"math/rand"
	"time"
)

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

func (area *BoardArea) resolveCombatPvP() bool {
	defense := area.Unit != nil
	return defense
}

func (area *BoardArea) resolveCombatPvE() {
	order := getOnlyOrder(area.IncomingMoves)

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
		order.succeedMove()
	} else {
		order.failMove()
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

func RollDice() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(6) + 1
}
