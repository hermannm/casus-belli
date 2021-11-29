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
	board.crossDangerZones()
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

func (board Board) crossDangerZones() {
	for _, area := range board {
		if area.Outgoing == nil || area.Outgoing.Type != Move {
			continue
		}

		move := area.Outgoing

		if destination, ok := area.GetNeighbor(move.To.Name, move.Via); ok {
			if destination.DangerZone != "" {
				move.crossDangerZone()
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
	resolved := make(map[string]bool)

	for !allResolved {
		allResolved = true

		for _, area := range board {
			if resolved[area.Name] {
				continue
			}

			if area.Unit != nil || len(area.IncomingMoves) != 1 {
				continue
			}

			move := area.IncomingMoves[0]
			if !area.HasNeighbor(move.To.Name) {
				transported := move.Transport()

				if !transported {
					resolved[area.Name] = true
					continue
				}
			}

			allResolved = false

			if area.Control == Uncontrolled {
				area.resolveCombatPvE()
			} else {
				move.succeedMove()
			}
			resolved[area.Name] = true
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

		area.resolveCombat()
	}
}

func (board Board) resolveBorderConflicts() {
	processed := make(map[string]bool)

	for _, area1 := range board {
		if area1.Outgoing == nil ||
			area1.Outgoing.Type != Move ||
			processed[area1.Name] {

			continue
		}

		area2 := area1.Outgoing.To

		processed[area1.Name], processed[area2.Name] = true, true

		if area2.Outgoing == nil ||
			area2.Outgoing.Type != Move ||
			area1.Name != area2.Outgoing.To.Name ||
			processed[area2.Name] {

			continue
		}

		if !area1.HasNeighbor(area2.Name) {
			success1 := area1.Outgoing.Transport()
			success2 := area2.Outgoing.Transport()

			if !success1 || !success2 {
				continue
			}
		}

		resolveBorderCombat(area1, area2)
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

			for _, move := range area.IncomingMoves {
				if !move.From.HasNeighbor(move.To.Name) {
					move.Transport()
				}
			}

			if area.Outgoing != nil && area.Outgoing.Type == Move {
				allResolved = false
				continue
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
