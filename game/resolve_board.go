package game

// Applies changes to the board given a round of orders.
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

// Resolves moves on the board in order.
func (board Board) resolveMoves() {
	board.crossDangerZones()
	board.cutSupports()
	board.resolveConflictFreeOrders()
	board.resolveTransportOrders()
	board.resolveBorderConflicts()
	board.resolveConflicts()
}

// Takes a list of orders and adds references to them in the board's areas.
func (board Board) populateAreaOrders(orders []*Order) {
	for _, order := range orders {
		if from, ok := board[order.From.Name]; ok {
			from.Order = order
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

// Finds move orders attempting to cross danger zones to their destinations,
// and checks if they fail.
func (board Board) crossDangerZones() {
	for _, area := range board {
		if area.Order == nil || area.Order.Type != Move {
			continue
		}

		move := area.Order

		destination, adjacent := area.GetNeighbor(move.To.Name, move.Via)
		if adjacent && destination.DangerZone != "" {
			move.crossDangerZone()
		}
	}
}

// Removes support orders that are attacked.
// If not attacked, checks if support is across danger zone, and if it fails.
func (board Board) cutSupports() {
	for _, area := range board {
		if area.Order == nil || area.Order.Type != Support {
			continue
		}

		support := area.Order

		if len(area.IncomingMoves) > 0 {
			support.failSupport()
			continue
		}

		destination, adjacent := area.GetNeighbor(support.To.Name, support.Via)
		if adjacent && destination.DangerZone != "" {
			support.crossDangerZone()
		}
	}
}

// Goes through areas that can be resolved without PvP combat, and resolves them.
func (board Board) resolveConflictFreeOrders() {
	allResolved := false
	processed := make(map[string]bool)

	// Keeps looping to potentially discover orders that can be resolved after others.
	for !allResolved {
		allResolved = true

		for _, area := range board {
			if processed[area.Name] || !area.IsEmpty() {
				continue
			}

			if len(area.IncomingMoves) != 1 {
				processed[area.Name] = true
				continue
			}

			move := area.IncomingMoves[0]

			// Checks if transport-dependent order can be transported without combat.
			// If it cannot, adds it to 'resolved' map to avoid repeating Transportable calculation.
			if !area.HasNeighbor(move.To.Name) {
				transportable, dangerZone := move.Transportable()

				if !transportable {
					processed[area.Name] = true
					continue
				}

				if dangerZone && !move.crossDangerZone() {
					processed[area.Name] = true
					continue
				}
			}

			allResolved = false

			if area.Control == Uncontrolled {
				area.resolveCombatPvE()
			} else {
				move.succeedMove()
			}
			processed[area.Name] = true
		}
	}
}

// Finds transport orders under attack, and resolves their combat.
func (board Board) resolveTransportOrders() {
	for _, area := range board {
		if area.Order != nil &&
			area.Order.Type == Transport &&
			len(area.IncomingMoves) > 0 {

			area.resolveCombat()
		}
	}
}

// Finds pairs of areas on the board that are attacking each other,
// and resolves combat between them.
func (board Board) resolveBorderConflicts() {
	processed := make(map[string]bool)

	for _, area1 := range board {
		if area1.Order == nil ||
			area1.Order.Type != Move ||
			processed[area1.Name] {

			continue
		}

		area2 := area1.Order.To

		if area2.Order == nil ||
			area2.Order.Type != Move ||
			area1.Name != area2.Order.To.Name ||
			processed[area2.Name] {

			continue
		}

		processed[area1.Name], processed[area2.Name] = true, true

		// If attacks must transport, both must succeed transport for this to still be a border conflict.
		if !area1.HasNeighbor(area2.Name) {
			success1 := area1.Order.Transport()
			success2 := area2.Order.Transport()

			if !success1 || !success2 {
				continue
			}
		}

		resolveBorderCombat(area1, area2)
	}
}

// Goes through areas that could not be previously resolved due to conflicting orders,
// and resolves them.
func (board Board) resolveConflicts() {
	allResolved := false
	processed := make(map[string]bool)

	// Keeps looping to potentially discover orders that can be resolved after others.
	for !allResolved {
		allResolved = true

		for _, area := range board {
			if processed[area.Name] {
				continue
			}

			for _, move := range area.IncomingMoves {
				if !move.From.HasNeighbor(move.To.Name) {
					move.Transport()
				}
			}

			if len(area.IncomingMoves) == 0 {
				processed[area.Name] = true
				continue
			}

			if area.Order != nil && area.Order.Type == Move {
				allResolved = false
				continue
			}

			area.resolveCombat()
			processed[area.Name] = true
		}
	}
}

// Goes through areas with siege orders, and updates the area following the siege.
func (board Board) resolveSieges() {
	for _, area := range board {
		if area.Order == nil || area.Order.Type != Besiege {
			continue
		}

		area.SiegeCount++
		area.Order.Status = Success
		area.Order = nil

		if area.SiegeCount == 2 {
			area.Control = area.Unit.Player
			area.SiegeCount = 0
		}
	}
}

// Cleans up remaining order references on the board after the round.
func (board Board) cleanup() {
	for _, area := range board {
		if area.Order != nil {
			area.Order.Status = Success
			area.Order = nil
		}

		if len(area.IncomingSupports) > 0 {
			area.IncomingSupports = make([]*Order, 0)
		}
	}
}

// Goes through the board and resolves winter orders.
// TODO: Implement.
func (board Board) resolveWinter(orders []*Order) {

}
