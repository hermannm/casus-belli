package game

func (order *Order) failMove() {
	order.Status = Fail
	order.From.Outgoing = nil
	delete(order.To.IncomingMoves, order.From.Name)
}

func (order *Order) succeedMove() {
	order.To.Control = order.Player.Color
	order.To.Unit = order.From.Unit
	order.Status = Success
	order.From.Unit = nil
	order.From.Outgoing = nil
	delete(order.To.IncomingMoves, order.From.Name)
}

func (order *Order) winCombat() {
	if order.To.Outgoing != nil {
		order.To.Outgoing.Status = Fail
	}

	for _, move := range order.To.IncomingMoves {
		if move == order {
			move.succeedMove()
		} else {
			move.failMove()
		}
	}
}

func (area *BoardArea) failTransportDependentMoves() {
	transportNeighbors := area.transportNeighbors(make(map[string]*BoardArea))

	for _, area := range transportNeighbors {
		for from, move := range area.IncomingMoves {
			if _, ok := area.Neighbors[from]; !ok {
				if !move.Transportable() {
					move.failMove()
				}
			}
		}
	}
}

func getOnlyOrder(orders map[string]*Order) *Order {
	for _, order := range orders {
		return order
	}
	return nil
}
