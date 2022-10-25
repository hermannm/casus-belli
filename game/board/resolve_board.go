package board

import "log"

// Adds the round's orders to the board, and resolves them.
// Returns a list of any potential battles from the round.
func (board Board) Resolve(round Round, messenger Messenger) (battles []Battle, winner string) {
	battles = make([]Battle, 0)

	switch round.Season {
	case SeasonWinter:
		board.resolveWinter(round.FirstOrders)
	default:
		firstBattles := board.resolveOrders(round.FirstOrders, messenger)
		battles = append(battles, firstBattles...)

		secondBattles := board.resolveOrders(round.SecondOrders, messenger)
		battles = append(battles, secondBattles...)

		board.resolveSieges()

		winner = board.resolveWinner()
	}

	board.cleanup()

	return battles, winner
}

// Resolves results of the given orders on the board.
func (board Board) resolveOrders(orders []Order, messenger Messenger) []Battle {
	battles := make([]Battle, 0)

	board.populateAreaOrders(orders)

	dangerZoneBattles := board.crossDangerZones()
	battles = append(battles, dangerZoneBattles...)
	err := messenger.SendBattleResults(dangerZoneBattles)
	if err != nil {
		log.Println(err)
	}

	singleplayerBattles := board.resolveMoves(false, messenger)
	battles = append(battles, singleplayerBattles...)

	remainingBattles := board.resolveMoves(true, messenger)
	battles = append(battles, remainingBattles...)

	return battles
}

// Takes a list of orders, and populates the appropriate areas on the board with those orders.
// Does not add support orders that have moves against them, as that cancels them.
func (board Board) populateAreaOrders(orders []Order) {
	// First adds all orders except supports, so that supports can check IncomingMoves.
	for _, order := range orders {
		if order.Type == OrderSupport {
			continue
		}

		board.addOrder(order)
	}

	// Then adds all supports, except in those areas that are attacked.
	for _, order := range orders {
		if order.Type != OrderSupport || len(board.Areas[order.From].IncomingMoves) > 0 {
			continue
		}

		board.addOrder(order)
	}
}

// Resolves moves on the board. Returns any resulting battles.
// Only resolves battles between players if allowPlayerConflict is true.
func (board Board) resolveMoves(allowPlayerConflict bool, messenger Messenger) []Battle {
	battles := make([]Battle, 0)

	battleReceiver := make(chan Battle)
	processing := make(map[string]struct{})
	processed := make(map[string]struct{})
	retreats := make(map[string]Order)

OuterLoop:
	for {
		select {
		case battle := <-battleReceiver:
			battles = append(battles, battle)
			messenger.SendBattleResults([]Battle{battle})

			newRetreats := board.resolveBattle(battle)
			for _, retreat := range newRetreats {
				retreats[retreat.From] = retreat
			}

			for _, area := range battle.areaNames() {
				delete(processing, area)
			}
		default:
		BoardLoop:
			for areaName, area := range board.Areas {
				retreat, hasRetreat := retreats[areaName]

				_, isProcessed := processed[areaName]
				if isProcessed && !hasRetreat {
					continue BoardLoop
				}

				_, isProcessing := processing[areaName]
				if isProcessing {
					continue BoardLoop
				}

				for _, move := range area.IncomingMoves {
					transportAttacked, dangerZones := board.resolveTransports(move, area)

					if transportAttacked {
						if allowPlayerConflict {
							continue BoardLoop
						} else {
							processed[areaName] = struct{}{}
						}
					} else if len(dangerZones) > 0 {
						survived, dangerZoneCrossings := move.crossDangerZones(dangerZones)
						if !survived {
							board.removeMove(move)
						}

						battles = append(battles, dangerZoneCrossings...)
						err := messenger.SendBattleResults(dangerZoneCrossings)
						if err != nil {
							log.Println(err)
						}
					}
				}

				moveCount := len(area.IncomingMoves)
				if moveCount == 0 {
					if hasRetreat && area.IsEmpty() {
						board.Areas[areaName] = area.setUnit(retreat.Unit)
						delete(retreats, areaName)
					}

					processed[area.Name] = struct{}{}
					continue BoardLoop
				}

				board.resolveAreaMoves(
					area,
					moveCount,
					allowPlayerConflict,
					battleReceiver,
					processing,
					processed,
					messenger,
				)
			}

			if len(processing) == 0 && len(retreats) == 0 {
				break OuterLoop
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

	for areaName, area := range board.Areas {
		order := area.Order

		if order.Type != OrderMove && order.Type != OrderSupport {
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
			if order.Type == OrderMove {
				board.Areas[areaName] = area.setUnit(Unit{})
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
	for areaName, area := range board.Areas {
		if area.Order.IsNone() || area.Order.Type != OrderBesiege {
			continue
		}

		area.SiegeCount++
		if area.SiegeCount == 2 {
			area.ControllingPlayer = area.Unit.Player
			area.SiegeCount = 0
		}

		board.Areas[areaName] = area
	}
}

// Goes through the board to check if any player has met the board's winning castle count.
// If there is a winner, and there is no tie, returns the tag of that player.
// Otherwise, returns "".
func (board Board) resolveWinner() string {
	castleCount := make(map[string]int)

	for _, area := range board.Areas {
		if area.Castle && area.IsControlled() {
			castleCount[area.ControllingPlayer]++
		}
	}

	tie := false
	highestCount := 0
	var highestCountPlayer string
	for player, count := range castleCount {
		if count > highestCount {
			highestCount = count
			highestCountPlayer = player
			tie = false
		} else if count == highestCount {
			tie = true
		}
	}

	if !tie && highestCount > board.WinningCastleCount {
		return highestCountPlayer
	}

	return ""
}

// Resolves winter orders (builds and internal moves) on the board. Assumes they have already been validated.
func (board Board) resolveWinter(orders []Order) {
	for _, order := range orders {
		switch order.Type {
		case OrderBuild:
			from := board.Areas[order.From]
			from.Unit = Unit{
				Player: order.Player,
				Type:   order.Build,
			}
			board.Areas[order.From] = from
		case OrderMove:
			from := board.Areas[order.From]
			to := board.Areas[order.To]

			to.Unit = from.Unit
			from.Unit = Unit{}

			board.Areas[order.From] = from
			board.Areas[order.To] = to
		}
	}
}

// Cleans up remaining order references on the board after the round.
func (board Board) cleanup() {
	for areaName, area := range board.Areas {
		area.Order = Order{}

		if len(area.IncomingMoves) > 0 {
			area.IncomingMoves = make([]Order, 0)
		}
		if len(area.IncomingSupports) > 0 {
			area.IncomingSupports = make([]Order, 0)
		}

		board.Areas[areaName] = area
	}
}
