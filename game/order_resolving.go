package game

func (board Board) Resolve(round *Round) {
	switch round.Season {
	case Winter:
		board.resolveWinter(round.FirstOrders)
	default:
		board.populateAreaOrders(round.FirstOrders)
		board.resolveMoves()
		board.populateAreaOrders(round.SecondOrders)
		board.resolveMoves()
		board.resolveSieges()
		board.cleanup()
	}
}

func (board Board) resolveMoves() {
	board.cutSupports()

	board.resolveConflictFreeOrders()

	board.resolveTransportOrders()
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
		if area.Outgoing != nil && area.Outgoing.Type != Transport {
			continue
		}

		if len(area.IncomingMoves) == 0 {
			continue
		}

		area.resolveCombat()

		if area.Outgoing == nil {
			area.failTransportDependentMoves()
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

func (board Board) resolveSieges() {
	for _, area := range board {
		if area.Outgoing != nil && area.Outgoing.Type == Besiege {
			area.SiegeCount++
			area.Outgoing.Status = Success
			area.Outgoing = nil
			if area.SiegeCount == 2 {
				area.Control = area.Unit.Color
				area.SiegeCount = 0
			}
		}
	}
}

func (board Board) cleanup() {
	for _, area := range board {
		if area.Outgoing != nil {
			area.Outgoing.Status = Success
			area.Outgoing = nil
		}

		if len(area.IncomingSupports) > 0 {
			area.IncomingSupports = make(map[string]*Order)
		}
	}
}

func (board Board) resolveWinter(orders []*Order) {

}
