package game

func ResolveRound(round *Round) {
	activeOrders := resolveOrders(round.Board, round.FirstOrders)
	round.SecondOrders = append(round.SecondOrders, activeOrders...)
	activeOrders = resolveOrders(round.Board, round.SecondOrders)
	resolveFinalOrders(round.Board, activeOrders)
}

func populateAreaOrders(board Board, orders []*Order) {
	for _, order := range orders {
		if to, ok := board[order.To.Name]; ok {
			switch order.Type {
			case Move:
				to.IncomingMoves[order.From.Name] = order
			case Support:
				to.IncomingSupports[order.From.Name] = order
			}
		}
		if from, ok := board[order.From.Name]; ok {
			from.Outgoing = order
		}
	}
}

func cutSupports(board Board) {
	for _, area := range board {
		if area.Outgoing.Type == Support {
			if len(area.IncomingMoves) > 0 {
				area.Outgoing.Status = Fail
				delete(area.Outgoing.To.IncomingSupports, area.Outgoing.From.Name)
			}
		}
	}
}

func resolveOrders(board Board, orders []*Order) []*Order {
	populateAreaOrders(board, orders)
	cutSupports(board)

	conflictFreeResolved := false
	for !conflictFreeResolved {
		conflictFreeResolved = resolveConflictFreeOrders(board)
	}

	resolveTransportOrders(board)

	activeOrders := []*Order{}
	for _, order := range orders {
		if order.Status == Pending {
			activeOrders = append(activeOrders, order)
		}
	}
	return activeOrders
}

func resolveConflictFreeOrders(board Board) bool {
	allResolved := true

	for _, area := range board {
		if area.Unit != nil || len(area.IncomingMoves) != 1 {
			continue
		}

		allResolved = false

		if area.Control == Uncontrolled {
			resolveCombatPvE(area)
		} else {
			succeedMove(area, getOnlyOrder(area.IncomingMoves))
		}
	}

	return allResolved
}

func resolveTransportOrders(board Board) {
	for _, area := range board {
		if area.Outgoing.Type != Transport {
			continue
		}

		if len(area.IncomingMoves) == 0 {
			continue
		}

		resolveCombat(area)

		if area.Outgoing == nil {
			failTransportDependentMoves(area)
		}
	}
}

func failTransportDependentMoves(area *BoardArea) {
	transportNeighbors := findTransportNeighbors(area, make(map[string]*BoardArea))

	for _, area := range transportNeighbors {
		for from, move := range area.IncomingMoves {
			if _, ok := area.Neighbors[from]; !ok {
				if !Transportable(move) {
					failMove(move)
				}
			}
		}

	}
}

func resolveFinalOrders(board Board, orders []*Order) {

}
