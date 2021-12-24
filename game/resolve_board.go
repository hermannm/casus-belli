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
	board.resolveMoveCycles()
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
				move.moveAndSucceed()
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

type cycleState struct {
	unit   Unit
	order  *Order
	combat bool
	winner Player
	tie    bool
}

// Resolves cycles of move orders (more than 2 move orders going in circle).
func (board Board) resolveMoveCycles() {
	processed := make(map[string]bool)

	for name, area := range board {
		if processed[name] {
			continue
		}

		cycle := area.discoverCycle(area.Name)

		if cycle == nil {
			continue
		}

		cycleStates := make(map[string]cycleState)

		for _, cycleArea := range cycle {
			processed[cycleArea.Order.To.Name] = true

			cycleStates[cycleArea.Order.To.Name] = cycleState{
				unit:  cycleArea.Unit,
				order: cycleArea.Order,
			}

			if len(cycleArea.Order.To.IncomingMoves) < 2 {
				continue
			}

			winner, tie := cycleArea.Order.To.resolveCombatPvP(false)

			if state, ok := cycleStates[cycleArea.Order.To.Name]; ok {
				state.combat = true
				state.winner = winner
				state.tie = tie

				cycleStates[cycleArea.Order.To.Name] = state
			}
		}

		for _, cycleArea := range cycle {
			state := cycleStates[cycleArea.Name]

			if !state.combat {
				if cycleArea.Control == Uncontrolled {
					win := cycleArea.calculateCombatPvE(state.order)

					if !win {
						state.order.failMove()
						continue
					}
				}

				cycleArea.removeUnit()
				state.order.succeedMove()

				cycleArea.Unit = state.unit
				if state.order.From.Unit == state.unit {
					state.order.From.Unit = Unit{}
				}

				state.order.From.Order = nil
				continue
			}

			// Ties already handled by resolveCombatPvP.
			if state.tie {
				continue
			}

			for _, move := range area.IncomingMoves {
				if move.Player != state.winner {
					move.failMove()
					move.killAttacker()
					continue
				}

				if cycleArea.Control == Uncontrolled {
					win := cycleArea.calculateCombatPvE(state.order)

					if !win {
						state.order.failMove()
						continue
					}
				}

				if move.Player != state.unit.Player {
					move.moveAndSucceed()
					continue
				}

				move.succeedMove()
				cycleArea.Unit = state.unit
				if state.order.From.Unit == state.unit {
					state.order.From.Unit = Unit{}
				}
			}
		}
	}
}

// Recursively finds a cycle of moves starting and ending with the given firstArea name.
// Assumes that border conflicts (move cycles with just 2 areas) are already solved.
// Returns a list of pointers to the areas in the cycle, or nil if no cycle was found.
func (area *BoardArea) discoverCycle(firstArea string) []*BoardArea {
	if area.Order == nil || area.Order.Type != Move {
		return nil
	}

	// The base case: the destination is the beginning of the cycle.
	if area.Order.To.Name == firstArea {
		return []*BoardArea{area}
	}

	// If the base case is not yet reached, pass cycle discovery to the next area in the chain.
	continuation := area.Order.To.discoverCycle(firstArea)
	if continuation == nil {
		return nil
	} else {
		return append(continuation, area)
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
