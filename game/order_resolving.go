package game

func (round *Round) Resolve() {
	activeOrders := round.Board.resolve(round.FirstOrders)

	round.SecondOrders = append(round.SecondOrders, activeOrders...)
	activeOrders = round.Board.resolve(round.SecondOrders)

	resolveFinalOrders(round.Board, activeOrders)
}

func (board Board) resolve(orders []*Order) (stillActive []*Order) {
	board.populateAreaOrders(orders)

	board.cutSupports()

	board.resolveConflictFreeOrders()

	board.resolveTransportOrders()

	stillActive = make([]*Order, 0)
	for _, order := range orders {
		if order.Status == Pending {
			stillActive = append(stillActive, order)
		}
	}
	return stillActive
}

func (board Board) populateAreaOrders(orders []*Order) {
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

func (board Board) cutSupports() {
	for _, area := range board {
		if area.Outgoing != nil && area.Outgoing.Type == Support {
			if len(area.IncomingMoves) > 0 {
				area.Outgoing.Status = Fail
				delete(area.Outgoing.To.IncomingSupports, area.Outgoing.From.Name)
			}
		}
	}
}

func (board Board) resolveConflictFreeOrders() {
	allResolved := false

	for !allResolved {
		allResolved = true

		for _, area := range board {
			if area.Unit != nil || len(area.IncomingMoves) != 1 {
				continue
			}

			allResolved = false

			if area.Control == Uncontrolled {
				area.resolveCombatPvE()
			} else {
				getOnlyOrder(area.IncomingMoves).succeedMove()
			}
		}
	}
}

func (board Board) resolveTransportOrders() {
	for _, area := range board {
		if area.Outgoing.Type != Transport {
			continue
		}

		if len(area.IncomingMoves) == 0 {
			continue
		}

		area.resolveCombat()

		if area.Outgoing == nil {
			failTransportDependentMoves(area)
		}
	}
}

func failTransportDependentMoves(area *BoardArea) {
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

func resolveFinalOrders(board Board, orders []*Order) {

}
