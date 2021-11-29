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
	board.resolveBorderConflicts()
	board.resolveConflicts()
}

func (board Board) populateAreaOrders(orders []*Order) {
	for _, order := range orders {
		if from, ok := board[order.From.Name]; ok {
			from.Outgoing = order
		}

		if order.To == nil {
			continue
		}

		if to, ok := board[order.To.Name]; ok {
			switch order.Type {
			case Move:
				to.IncomingMoves = append(to.IncomingMoves, order)
			case Support:
				to.IncomingSupports = append(to.IncomingSupports, order)
			}
		}
	}
}

func (board Board) cutSupports() {
	for _, area := range board {
		if area.Outgoing != nil && area.Outgoing.Type == Support {
			if len(area.IncomingMoves) > 0 {
				area.Outgoing.Status = Fail

				area.Outgoing.To.IncomingSupports = removeOrder(
					area.Outgoing.To.IncomingSupports,
					area.Outgoing,
				)
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
				area.IncomingMoves[0].succeedMove()
			}
		}
	}
}

func (board Board) resolveTransportOrders() {
	for _, area := range board {
		if area.Outgoing == nil || area.Outgoing.Type != Transport {
			continue
		}

		if len(area.IncomingMoves) == 0 {
			continue
		}

		player := area.Unit.Color

		area.resolveCombat()

		if area.Outgoing == nil {
			area.failTransportDependencies(player, make(map[string]bool))
		}
	}
}

func (board Board) resolveBorderConflicts() {
	processed := make(map[string]bool)

	for name, area := range board {
		if area.Outgoing != nil &&
			area.Outgoing.Type == Move &&
			area.Outgoing.To.Outgoing != nil &&
			area.Outgoing.To.Outgoing.Type == Move &&
			name == area.Outgoing.To.Outgoing.To.Name {

			area2 := area.Outgoing.To

			_, processedArea1 := processed[name]
			_, processedArea2 := processed[area2.Name]

			if !processedArea1 && !processedArea2 {
				processed[name] = true
				processed[area2.Name] = true

				resolveBorderCombat(area, area2)
			}
		}
	}
}

func (board Board) resolveConflicts() {
	allResolved := false

	for !allResolved {
		allResolved = true

		for _, area := range board {
			if len(area.IncomingMoves) == 0 {
				continue
			}

			if area.Outgoing != nil && area.Outgoing.Type == Move {
				allResolved = false
			}

			area.resolveCombat()
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
			area.IncomingSupports = make([]*Order, 0)
		}
	}
}

func (board Board) resolveWinter(orders []*Order) {

}
