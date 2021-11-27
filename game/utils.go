package game

func failMove(order *Order) {
	order.Status = Fail
	order.From.Outgoing = nil
	delete(order.To.IncomingMoves, order.From.Name)
}

func succeedMove(area *BoardArea, order *Order) {
	for _, move := range area.IncomingMoves {
		if move == order {
			area.Control = order.Player.Color
			area.Unit = order.From.Unit
			order.Status = Success
			order.From.Unit = nil
			order.From.Outgoing = nil
			delete(area.IncomingMoves, order.From.Name)
		} else {
			failMove(move)
		}
	}
}

func getOnlyOrder(orders map[string]*Order) *Order {
	for _, order := range orders {
		return order
	}
	return nil
}

func mergeMaps(maps ...map[string]*BoardArea) map[string]*BoardArea {
	newMap := make(map[string]*BoardArea)

	for _, subMap := range maps {
		for key, area := range subMap {
			newMap[key] = area
		}
	}

	return newMap
}

func copyMap(oldMap map[string]*BoardArea) map[string]*BoardArea {
	newMap := make(map[string]*BoardArea)
	for key, area := range oldMap {
		newMap[key] = area
	}
	return newMap
}
