package game

import (
	. "immerse-ntnu/hermannia/server/types"
)

func resolveCombat(area *BoardArea) {
	moveOrders := []*Order{}
	supportOrders := []*Order{}

	for _, order := range area.Incoming {
		switch order.Type {
		case Move:
			moveOrders = append(moveOrders, order)
		case Support:
			supportOrders = append(supportOrders, order)
		}
	}

	if len(moveOrders) == 1 {
		order := moveOrders[0]
		attackMods := AttackModifiers(*order, false)

		if area.Control == Uncontrolled {
			mods := append(attackMods, Modifier{
				Type:  DiceMod,
				Value: RollDice(),
			})

			for _, supportOrder := range supportOrders {
				if callSupport(*order, *supportOrder) {
					mods = append(mods, Modifier{
						Type:  SupportMod,
						Value: 1,
					})
				}
			}

			total := 0
			for _, mod := range mods {
				total += mod.Value
			}

			if total <= 4 {
				succeedMove(order)
				order.Result = CombatResult{
					Total: total,
					Parts: mods,
				}
			}
		}
	}
}

func callSupport(moveOrder Order, supportOrder Order) bool {
	if moveOrder.Player.Color == supportOrder.Player.Color {
		return true
	}

	return false
}
