package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

// Finds move and support orders attempting to cross danger zones to their destinations, and fails
// them if they don't make it across.
func resolveDangerZones(board gametypes.Board) (results []gametypes.Battle) {
	for regionName, region := range board.Regions {
		order := region.Order

		if order.Type != gametypes.OrderMove && order.Type != gametypes.OrderSupport {
			continue
		}

		destination, adjacent := region.GetNeighbor(order.Destination, order.ViaDangerZone)
		if !adjacent || destination.DangerZone == "" {
			continue
		}

		survived, result := crossDangerZone(order, destination.DangerZone)
		results = append(results, result)

		if !survived {
			if order.Type == gametypes.OrderMove {
				region.Unit = gametypes.Unit{}
				board.Regions[regionName] = region
			}

			board.RemoveOrder(order)
		}
	}

	return results
}

// Rolls dice to see if order makes it across danger zone.
func crossDangerZone(
	order gametypes.Order,
	dangerZone string,
) (survived bool, result gametypes.Battle) {
	diceMod := gametypes.RollDiceBonus()

	result = gametypes.Battle{
		Results: []gametypes.Result{
			{Total: diceMod.Value, Parts: []gametypes.Modifier{diceMod}, Move: order},
		},
		DangerZone: dangerZone,
	}

	// All danger zones currently require a dice roll greater than 2.
	// May need to be changed in the future if a more dynamic implementation is preferred.
	return diceMod.Value > 2, result
}

func crossDangerZones(
	order gametypes.Order,
	dangerZones []string,
) (survivedAll bool, results []gametypes.Battle) {
	for _, dangerZone := range dangerZones {
		survived, result := crossDangerZone(order, dangerZone)
		results = append(results, result)
		if !survived {
			return false, results
		}
	}

	return true, results
}
