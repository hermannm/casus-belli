package game

// Finds support orders attempting to cross danger zones to their destinations, and fails them if
// they don't make it across.
func (game *Game) resolveDangerZoneSupports() {
	var results []Battle
	for _, region := range game.board {
		results = game.board.resolveDangerZoneCrossings(region, region.incomingSupports, results)
	}

	if len(results) != 0 {
		if err := game.messenger.SendBattleResults(results...); err != nil {
			game.log.Error(err)
		}
	}
}

// Finds incoming move orders to the given region that attempt to cross danger zones, and kills them
// if they fail.
func (game *Game) resolveDangerZoneMoves(region *Region) {
	results := game.board.resolveDangerZoneCrossings(region, region.incomingMoves, nil)
	if len(results) != 0 {
		if err := game.messenger.SendBattleResults(results...); err != nil {
			game.log.Error(err)
		}
	}

	region.dangerZonesResolved = true
}

func (board Board) resolveDangerZoneCrossings(
	region *Region,
	incomingMovesOrSupports []Order,
	resultsToAppend []Battle,
) []Battle {
	for _, order := range incomingMovesOrSupports {
		crossing, adjacent := region.getNeighbor(order.Origin, order.ViaDangerZone)

		// Non-adjacent moves are handled by resolveTransports
		if !adjacent || crossing.DangerZone == "" {
			continue
		}

		survived, result := crossDangerZone(order, crossing.DangerZone)
		resultsToAppend = append(resultsToAppend, result)

		if !survived {
			if order.Type == OrderMove {
				board.killMove(order)
			} else {
				board.removeOrder(order)
			}
		}
	}

	return resultsToAppend
}

// Rolls dice to see if order makes it across danger zone.
func crossDangerZone(order Order, dangerZone DangerZone) (survived bool, result Battle) {
	diceModifier := Modifier{Type: ModifierDice, Value: rollDice()}

	result = Battle{
		Results: []Result{
			{Total: diceModifier.Value, Parts: []Modifier{diceModifier}, Move: order},
		},
		DangerZone: dangerZone,
	}

	// All danger zones currently require a dice roll greater than 2.
	// May need to be changed in the future if a more dynamic implementation is preferred.
	return diceModifier.Value > 2, result
}

func crossDangerZones(
	order Order,
	dangerZones []DangerZone,
) (survivedAll bool, results []Battle) {
	for _, dangerZone := range dangerZones {
		survived, result := crossDangerZone(order, dangerZone)
		results = append(results, result)
		if !survived {
			return false, results
		}
	}

	return true, results
}
