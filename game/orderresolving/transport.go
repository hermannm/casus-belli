package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

// Resolves transport of the given move to the given destination if it requires transport.
// If transported, returns whether the transport path is attacked,
// and a list of danger zones that the order must cross to transport, if any.
func resolveTransports(
	move gametypes.Order, destination gametypes.Region, board gametypes.Board,
) (transportAttacked bool, dangerZones []string) {
	if destination.HasNeighbor(move.From) {
		return false, nil
	}

	canTransport, transportAttacked, dangerZones := board.FindTransportPath(move.From, move.To)

	if !canTransport {
		board.RemoveOrder(move)
		return false, nil
	}

	return transportAttacked, dangerZones
}
