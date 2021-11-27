package game

func (order *Order) failMove() {
	order.Status = Fail
	order.From.Outgoing = nil
	delete(order.To.IncomingMoves, order.From.Name)
}

func (order *Order) succeedMove() {
	if order.To.Outgoing != nil {
		order.To.Outgoing.Status = Fail
	}

	for _, move := range order.To.IncomingMoves {
		if move == order {
			order.To.Control = order.Player.Color
			order.To.Unit = order.From.Unit
			order.Status = Success
			order.From.Unit = nil
			order.From.Outgoing = nil
			delete(order.To.IncomingMoves, order.From.Name)
		} else {
			move.failMove()
		}
	}
}

func getOnlyOrder(orders map[string]*Order) *Order {
	for _, order := range orders {
		return order
	}
	return nil
}
