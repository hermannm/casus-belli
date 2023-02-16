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

	board.populateRegionOrders(orders)

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

// Takes a list of orders, and populates the appropriate regions on the board with those orders.
// Does not add support orders that have moves against them, as that cancels them.
func (board Board) populateRegionOrders(orders []Order) {
	// First adds all orders except supports, so that supports can check IncomingMoves.
	for _, order := range orders {
		if order.Type == OrderSupport {
			continue
		}

		board.addOrder(order)
	}

	// Then adds all supports, except in those regions that are attacked.
	for _, order := range orders {
		if order.Type != OrderSupport || len(board.Regions[order.From].IncomingMoves) > 0 {
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

			for _, region := range battle.regionNames() {
				delete(processing, region)
			}
		default:
		BoardLoop:
			for regionName, region := range board.Regions {
				retreat, hasRetreat := retreats[regionName]

				_, isProcessed := processed[regionName]
				if isProcessed && !hasRetreat {
					continue BoardLoop
				}

				_, isProcessing := processing[regionName]
				if isProcessing {
					continue BoardLoop
				}

				for _, move := range region.IncomingMoves {
					transportAttacked, dangerZones := board.resolveTransports(move, region)

					if transportAttacked {
						if allowPlayerConflict {
							continue BoardLoop
						} else {
							processed[regionName] = struct{}{}
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

				moveCount := len(region.IncomingMoves)
				if moveCount == 0 {
					if hasRetreat && region.IsEmpty() {
						board.Regions[regionName] = region.setUnit(retreat.Unit)
						delete(retreats, regionName)
					}

					processed[region.Name] = struct{}{}
					continue BoardLoop
				}

				board.resolveRegionMoves(
					region,
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

	for regionName, region := range board.Regions {
		order := region.Order

		if order.Type != OrderMove && order.Type != OrderSupport {
			continue
		}

		// Checks if the order tries to cross a danger zone.
		destination, adjacent := region.GetNeighbor(order.To, order.Via)
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
				board.Regions[regionName] = region.setUnit(Unit{})
				board.removeMove(order)
			} else {
				board.removeSupport(order)
			}
		}
	}

	return battles
}

// Goes through regions with siege orders, and updates the region following the siege.
func (board Board) resolveSieges() {
	for regionName, region := range board.Regions {
		if region.Order.IsNone() || region.Order.Type != OrderBesiege {
			continue
		}

		region.SiegeCount++
		if region.SiegeCount == 2 {
			region.ControllingPlayer = region.Unit.Player
			region.SiegeCount = 0
		}

		board.Regions[regionName] = region
	}
}

// Goes through the board to check if any player has met the board's winning castle count.
// If there is a winner, and there is no tie, returns the tag of that player.
// Otherwise, returns "".
func (board Board) resolveWinner() string {
	castleCount := make(map[string]int)

	for _, region := range board.Regions {
		if region.Castle && region.IsControlled() {
			castleCount[region.ControllingPlayer]++
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

// Resolves winter orders (builds and internal moves) on the board.
// Assumes they have already been validated.
func (board Board) resolveWinter(orders []Order) {
	for _, order := range orders {
		switch order.Type {
		case OrderBuild:
			from := board.Regions[order.From]
			from.Unit = Unit{
				Player: order.Player,
				Type:   order.Build,
			}
			board.Regions[order.From] = from
		case OrderMove:
			from := board.Regions[order.From]
			to := board.Regions[order.To]

			to.Unit = from.Unit
			from.Unit = Unit{}

			board.Regions[order.From] = from
			board.Regions[order.To] = to
		}
	}
}

// Cleans up remaining order references on the board after the round.
func (board Board) cleanup() {
	for regionName, region := range board.Regions {
		region.Order = Order{}

		if len(region.IncomingMoves) > 0 {
			region.IncomingMoves = make([]Order, 0)
		}
		if len(region.IncomingSupports) > 0 {
			region.IncomingSupports = make([]Order, 0)
		}

		board.Regions[regionName] = region
	}
}
