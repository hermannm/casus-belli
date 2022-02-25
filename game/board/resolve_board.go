package board

// Adds the round's orders to the board, and resolves them.
// Returns a list of any potential battles from the round.
func (board Board) Resolve(round Round) []Battle {
	battles := make([]Battle, 0)

	switch round.Season {
	case Winter:
		board.resolveWinter(round.FirstOrders)
	default:
		firstBattles := board.resolveOrders(round.FirstOrders)
		battles = append(battles, firstBattles...)

		secondBattles := board.resolveOrders(round.SecondOrders)
		battles = append(battles, secondBattles...)

		board.resolveSieges()
	}

	board.cleanup()

	return battles
}

// Resolves results of the given orders on the board.
func (board Board) resolveOrders(orders []Order) []Battle {
	battles := make([]Battle, 0)

	board.populateAreaOrders(orders)

	dangerZoneBattles := board.crossDangerZones()
	battles = append(battles, dangerZoneBattles...)

	singleplayerBattles := board.resolveMoves(false)
	battles = append(battles, singleplayerBattles...)

	remainingBattles := board.resolveMoves(true)
	battles = append(battles, remainingBattles...)

	return battles
}

// Takes a list of orders, and populates the appropriate areas on the board with those orders.
// Does not add support orders that have moves against them, as that cancels them.
func (board Board) populateAreaOrders(orders []Order) {
	// First adds all orders except supports, so that supports can check IncomingMoves.
	for _, order := range orders {
		if order.Type == Support {
			continue
		}

		board.addOrder(order)
	}

	// Then adds all supports, except in those areas that are attacked.
	for _, order := range orders {
		if order.Type != Support || len(board[order.From].IncomingMoves) > 0 {
			continue
		}

		board.addOrder(order)
	}
}

// Resolves moves on the board. Returns any resulting battles.
// Only resolves battles between players if playerConflictsAllowed is true.
func (board Board) resolveMoves(playerConflictsAllowed bool) []Battle {
	battles := make([]Battle, 0)

	battleReceiver := make(chan Battle)
	processing := make(map[string]struct{})
	processed := make(map[string]struct{})
	retreats := make(map[string]Order)

outerLoop:
	for {
		select {
		case battle := <-battleReceiver:
			battles = append(battles, battle)

			newRetreats := board.resolveBattle(battle)
			for _, retreat := range newRetreats {
				retreats[retreat.From] = retreat
			}

			for _, area := range battle.areaNames() {
				delete(processing, area)
			}
		default:
		boardLoop:
			for areaName, area := range board {
				retreat, hasRetreat := retreats[areaName]
				if _, skip := processed[areaName]; skip && !hasRetreat {
					continue
				}
				if _, skip := processing[areaName]; skip {
					continue
				}
				if area.Order.Type == Move {
					continue
				}

				for _, move := range area.IncomingMoves {
					transportAttacked, dangerZoneCrossings := board.resolveTransports(move, area)

					if dangerZoneCrossings != nil {
						battles = append(battles, dangerZoneCrossings...)
					}

					if transportAttacked {
						if playerConflictsAllowed {
							continue boardLoop
						} else {
							processed[areaName] = struct{}{}
						}
					}
				}

				moveCount := len(area.IncomingMoves)
				if moveCount == 0 {
					processed[area.Name] = struct{}{}

					if hasRetreat && area.IsEmpty() {
						board[areaName] = area.setUnit(retreat.Unit)
						delete(retreats, areaName)
					}

					continue
				}

				board.resolveAreaMoves(
					area,
					moveCount,
					playerConflictsAllowed,
					battleReceiver,
					processing,
					processed,
				)
			}

			if len(processing) == 0 {
				if len(retreats) != 0 {
					continue
				}

				break outerLoop
			}
		}
	}

	return battles
}

// Finds move and support orders attempting to cross danger zones to their destinations,
// and fail them if they don't make it across.
// Returns a battle result for each danger zone crossing.
func (board Board) crossDangerZones() []Battle {
	battles := make([]Battle, 0)

	for areaName, area := range board {
		order := area.Order

		if order.Type != Move && order.Type != Support {
			continue
		}

		// Checks if the order tries to cross a danger zone.
		destination, adjacent := area.GetNeighbor(order.To, order.Via)
		if !adjacent || destination.DangerZone == "" {
			continue
		}

		// Resolves the danger zone crossing.
		survived, battle := order.crossDangerZone(destination.DangerZone)
		battles = append(battles, battle)

		// If move fails danger zone crossing, the unit dies.
		// If support fails crossing, only the order fails.
		if !survived {
			if order.Type == Move {
				board[areaName] = area.setUnit(Unit{})
				board.removeMove(order)
			} else {
				board.removeSupport(order)
			}
		}
	}

	return battles
}

// Goes through areas with siege orders, and updates the area following the siege.
func (board Board) resolveSieges() {
	for areaName, area := range board {
		if area.Order.IsNone() || area.Order.Type != Besiege {
			continue
		}

		area.SiegeCount++
		if area.SiegeCount == 2 {
			area.Control = area.Unit.Player
			area.SiegeCount = 0
		}

		board[areaName] = area
	}
}

// Resolves winter orders (builds and internal moves) on the board.
// Assumes they have already been validated.
func (board Board) resolveWinter(orders []Order) {
	for _, order := range orders {
		switch order.Type {

		case Build:
			from := board[order.From]
			from.Unit = Unit{
				Player: order.Player,
				Type:   order.Build,
			}
			board[order.From] = from

		case Move:
			from := board[order.From]
			to := board[order.To]

			to.Unit = from.Unit
			from.Unit = Unit{}

			board[order.From] = from
			board[order.To] = to

		}
	}
}

// Cleans up remaining order references on the board after the round.
func (board Board) cleanup() {
	for areaName, area := range board {
		area.Order = Order{}

		if len(area.IncomingMoves) > 0 {
			area.IncomingMoves = make([]Order, 0)
		}
		if len(area.IncomingSupports) > 0 {
			area.IncomingSupports = make([]Order, 0)
		}

		board[areaName] = area
	}
}
