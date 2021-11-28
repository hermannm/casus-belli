package game

func (order *Order) failMove() {
	order.Status = Fail
	order.From.Outgoing = nil
	delete(order.To.IncomingMoves, order.From.Name)
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
	delete(order.To.IncomingMoves, order.From.Name)
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
