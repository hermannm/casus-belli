package game

// Finds move and support orders attempting to cross danger zones to their destinations, and fails
// them if they don't make it across.
func resolveDangerZones(board Board) (results []Battle) {
	for regionName, region := range board {
		order := region.order

		if order.Type != OrderMove && order.Type != OrderSupport {
			continue
		}

		destination, adjacent := region.getNeighbor(order.Destination, order.ViaDangerZone)
		if !adjacent || destination.DangerZone == "" {
			continue
		}

		survived, result := crossDangerZone(order, destination.DangerZone)
		results = append(results, result)

		if !survived {
			if order.Type == OrderMove {
				region.Unit = Unit{}
				board[regionName] = region
			}

			board.removeOrder(order)
		}
	}

	return results
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
