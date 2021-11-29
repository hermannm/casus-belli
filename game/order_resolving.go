package game

func removeOrder(oldOrders []*Order, remove *Order) []*Order {
	newOrders := make([]*Order, 0)

	for _, order := range oldOrders {
		if order != remove {
			newOrders = append(newOrders, order)
		}
	}

	return newOrders
}

func (order *Order) failMove() {
	order.Status = Fail
	order.From.Outgoing = nil
	order.To.IncomingMoves = removeOrder(order.To.IncomingMoves, order)
}

func (order *Order) die() {
	order.From.Unit = nil
}

func (order *Order) succeedMove() {
	order.To.Control = order.Player
	order.To.Unit = order.From.Unit
	order.Status = Success
	order.From.Unit = nil
	order.From.Outgoing = nil
	order.To.IncomingMoves = removeOrder(order.To.IncomingMoves, order)
}

func (area *BoardArea) resolveWinner(winner PlayerColor) {
	if area.Outgoing != nil && area.Unit.Color != winner {
		area.Outgoing.Status = Fail
		area.Outgoing = nil
	}

	for _, move := range area.IncomingMoves {
		if move.Player == winner {
			move.succeedMove()
		} else {
			move.failMove()
			move.die()
		}
	}
}

func (order *Order) crossDangerZone() bool {
	diceMod := DiceModifier()

	combat := Combat{
		{
			Total:  diceMod.Value,
			Parts:  []Modifier{diceMod},
			Player: order.Player,
		},
	}

	order.From.Combats = append(order.From.Combats, combat)

	if diceMod.Value <= 2 {
		order.failMove()
		order.die()
		return false
	}
	return true
}
