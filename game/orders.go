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
	order.To.Control = order.Player.Color
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
		if move.Player.Color == winner {
			move.succeedMove()
		} else {
			move.failMove()
			move.die()
		}
	}
}

func (area *BoardArea) failTransportDependentMoves() {
	transportNeighbors, _ := area.TransportNeighbors(make([]*BoardArea, 0))

	for _, area := range transportNeighbors {
		for _, move := range area.IncomingMoves {
			if !area.HasNeighbor(move.From.Name) {
				if _, transportable := transportNeighbors[move.To.Name]; !transportable {
					move.failMove()
				}
			}
		}
	}
}
