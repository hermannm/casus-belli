package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

// Moves the unit of the given move order to its destination,
// killing any unit that may have already been there,
// and sets control of the region to the order's player.
//
// Then removes references to this move on the board,
// and removes any potential order from the destination region.
func succeedMove(move gametypes.Order, board gametypes.Board) {
	to := board.Regions[move.To]
	to.Unit = move.Unit
	to.Order = gametypes.Order{}
	if !to.Sea {
		to.ControllingPlayer = move.Player
	}
	board.Regions[move.To] = to

	board.RemoveUnit(move.Unit, move.From)

	board.RemoveOrder(move)
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

// Attempts to move the unit of the given move order back to its origin.
// Returns whether the retreat succeeded.
func attemptRetreat(move gametypes.Order, board gametypes.Board) bool {
	from := board.Regions[move.From]

	if from.Unit == move.Unit {
		return true
	}

	if len(from.IncomingMoves) != 0 {
		return false
	}

	from.Unit = move.Unit
	board.Regions[move.From] = from
	return true
}
