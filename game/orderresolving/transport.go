package orderresolving

import (
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
)

// Resolves transport of the given move to its destination, if it requires transport.
// If the transport depends on other orders to resolve first, returns transportMustWait=true.
func resolveTransport(
	move gametypes.Order, board gametypes.Board, resolverState *ResolverState, messenger Messenger,
) (transportMustWait bool) {
	// If the move is between two adjacent regions, then it does not need transport.
	if board.Regions[move.Destination].HasNeighbor(move.Origin) {
		return false
	}

	canTransport, transportAttacked, dangerZones := board.FindTransportPath(move.Origin, move.Destination)

	if !canTransport {
		board.RemoveOrder(move)
		return false
	}

	if transportAttacked {
		if resolverState.allowPlayerConflict {
			resolverState.processed.Add(move.Destination)
		}

		return true
	}

	if len(dangerZones) > 0 {
		survived, dangerZoneBattles := crossDangerZones(move, dangerZones)

		if !survived {
			board.RemoveOrder(move)
		}

		resolverState.resolvedBattles = append(resolverState.resolvedBattles, dangerZoneBattles...)
		if err := messenger.SendBattleResults(dangerZoneBattles); err != nil {
			log.Println(err)
		}

		return false
	}

	return false
}
