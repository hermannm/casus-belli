package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

// Finds move and support orders attempting to cross danger zones to their destinations,
// and fail them if they don't make it across.
// Returns a battle result for each danger zone crossing.
func resolveDangerZones(board gametypes.Board) []gametypes.Battle {
	battles := make([]gametypes.Battle, 0)

	for regionName, region := range board.Regions {
		order := region.Order

		if order.Type != gametypes.OrderMove && order.Type != gametypes.OrderSupport {
			continue
		}

		// Checks if the order tries to cross a danger zone.
		destination, adjacent := region.GetNeighbor(order.Destination, order.Via)
		if !adjacent || destination.DangerZone == "" {
			continue
		}

		// Resolves the danger zone crossing.
		survived, battle := crossDangerZone(order, destination.DangerZone)
		battles = append(battles, battle)

		// If move fails danger zone crossing, the unit dies.
		// If support fails crossing, only the order fails.
		if !survived {
			if order.Type == gametypes.OrderMove {
				region.Unit = gametypes.Unit{}
				board.Regions[regionName] = region
			}

			board.RemoveOrder(order)
		}
	}

	return battles
}

// Rolls dice to see if order makes it across danger zone.
// Returns whether the order succeeded, and the resulting battle for use by the client.
func crossDangerZone(
	order gametypes.Order, dangerZone string,
) (survived bool, result gametypes.Battle) {
	diceMod := gametypes.RollDiceBonus()

	// Records crossing attempt as a battle, so clients can see dice roll.
	battle := gametypes.Battle{
		Results: []gametypes.Result{
			{Total: diceMod.Value, Parts: []gametypes.Modifier{diceMod}, Move: order},
		},
		DangerZone: dangerZone,
	}

	// All danger zones currently require a dice roll greater than 2.
	// May need to be changed in the future if a more dynamic implementation is preferred.
	return diceMod.Value > 2, battle
}

func crossDangerZones(
	order gametypes.Order, dangerZones []string,
) (survivedAll bool, results []gametypes.Battle) {
	survivedAll = true
	results = make([]gametypes.Battle, 0)

	for _, dangerZone := range dangerZones {
		survived, result := crossDangerZone(order, dangerZone)
		results = append(results, result)
		if !survived {
			survivedAll = false
		}
	}

	return survivedAll, results
}
