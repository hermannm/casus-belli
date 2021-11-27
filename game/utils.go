package game

func removeOrder(orders []*Order, orderToRemove *Order) []*Order {
	newOrders := []*Order{}

	for _, order := range orders {
		if order != orderToRemove {
			newOrders = append(newOrders, order)
		}
	}

	return newOrders
}

func failMove(order *Order) {
	order.Status = Fail
	order.From.Outgoing = nil
}

func succeedMove(area *BoardArea, order *Order) {
	for _, move := range area.IncomingMoves {
		if move == order {
			area.Control = order.Player.Color
			area.Unit = order.From.Unit
			order.Status = Success
			order.From.Unit = nil
			order.From.Outgoing = nil
		} else {
			failMove(move)
		}
	}

	area.IncomingMoves = []*Order{}
}
