package game

import (
	"hermannm.dev/set"
)

func (game *Game) resolveTransport(move Order) (transportMustWait bool) {
	// If the move is between two adjacent regions, then it does not need transport
	if game.Board[move.Destination].hasNeighbor(move.Origin) {
		return false
	}

	canTransport, transportAttacked, dangerZones := game.Board.findTransportPath(
		move.Origin,
		move.Destination,
	)

	if !canTransport {
		game.Board.removeOrder(move)
		return false
	}

	if transportAttacked {
		return true
	}

	if len(dangerZones) > 0 {
		survived, dangerZoneBattles := crossDangerZones(move, dangerZones)

		if !survived {
			game.Board.removeOrder(move)
		}

		game.resolvedBattles = append(game.resolvedBattles, dangerZoneBattles...)
		if err := game.messenger.SendBattleResults(dangerZoneBattles...); err != nil {
			game.log.Error(err)
		}

		return false
	}

	return false
}

// Checks if a unit can be transported via ship from the given origin to the given destination.
func (board Board) findTransportPath(
	originName RegionName,
	destinationName RegionName,
) (canTransport bool, isTransportAttacked bool, dangerZones []DangerZone) {
	origin := board[originName]
	if origin.isEmpty() || origin.Unit.Type == UnitShip || origin.IsSea {
		return false, false, nil
	}

	return board.recursivelyFindTransportPath(origin, destinationName, &set.ArraySet[RegionName]{})
}

// Stores status of a path of transport orders to destination.
type transportPath struct {
	isAttacked  bool
	dangerZones []DangerZone
}

func (board Board) recursivelyFindTransportPath(
	region *Region,
	destination RegionName,
	regionsToExclude set.Set[RegionName],
) (canTransport bool, isTransportAttacked bool, dangerZones []DangerZone) {
	transportingNeighbors, newRegionsToExclude := region.getTransportingNeighbors(
		board,
		regionsToExclude,
	)

	var paths []transportPath

	for _, transportNeighbor := range transportingNeighbors {
		transportRegion := board[transportNeighbor.Name]

		destinationAdjacent, destinationDangerZone := region.checkNeighborsForDestination(
			destination,
		)

		// Recursively calls this function on the transporting neighbor,
		// in order to find potential transport chains.
		nextCanTransport, nextTransportAttacked, nextDangerZones := board.recursivelyFindTransportPath(
			transportRegion,
			destination,
			newRegionsToExclude,
		)

		var subPaths []transportPath
		if destinationAdjacent {
			subPaths = append(
				subPaths,
				transportPath{
					isAttacked:  transportRegion.isAttacked(),
					dangerZones: []DangerZone{destinationDangerZone},
				},
			)
		}
		if nextCanTransport {
			subPaths = append(
				subPaths,
				transportPath{
					isAttacked:  transportRegion.isAttacked() || nextTransportAttacked,
					dangerZones: nextDangerZones,
				},
			)
		}

		// If both this neighbor and potential subpaths can transport, finds the best one.
		// This is for the niche edge case of there being a danger zone between this
		// transport and the destination, in which case a longer subpath may be better.
		bestPath, canTransport := bestTransportPath(subPaths)
		if canTransport {
			paths = append(paths, bestPath)
		}
	}

	bestPath, canTransport := bestTransportPath(paths)
	return canTransport, bestPath.isAttacked, bestPath.dangerZones
}

func (region *Region) getTransportingNeighbors(
	board Board,
	regionsToExclude set.Set[RegionName],
) (transports []Neighbor, newRegionsToExclude set.Set[RegionName]) {
	newRegionsToExclude = regionsToExclude.Copy()

	if region.isEmpty() {
		return transports, newRegionsToExclude
	}

	for _, neighbor := range region.Neighbors {
		neighborRegion := board[neighbor.Name]

		if regionsToExclude.Contains(neighbor.Name) ||
			neighborRegion.order.Type != OrderTransport ||
			neighborRegion.Unit.Faction != region.Unit.Faction {
			continue
		}

		transports = append(transports, neighbor)
		newRegionsToExclude.Add(neighbor.Name)
	}

	return transports, newRegionsToExclude
}

func (region *Region) checkNeighborsForDestination(
	destination RegionName,
) (destinationIsAdjacent bool, mustGoThrough DangerZone) {
	for _, neighbor := range region.Neighbors {
		if neighbor.Name == destination {
			// If destination is already found to be adjacent but only through a danger zone,
			// checks if there is a different path to it without a danger zone.
			if destinationIsAdjacent {
				if mustGoThrough != "" {
					mustGoThrough = neighbor.DangerZone
				}
				continue
			}

			destinationIsAdjacent = true
			mustGoThrough = neighbor.DangerZone
		}
	}

	return destinationIsAdjacent, mustGoThrough
}

// Prioritizes paths that are not attacked first, then paths that have to cross the fewest danger
// zones.
func bestTransportPath(paths []transportPath) (bestPath transportPath, canTransport bool) {
	if len(paths) == 0 {
		return transportPath{}, false
	}

	bestPath = paths[0]

	for _, path := range paths[1:] {
		if bestPath.isAttacked {
			if path.isAttacked {
				if len(path.dangerZones) < len(bestPath.dangerZones) {
					bestPath = path
				}
			} else {
				bestPath = path
			}
		} else if !path.isAttacked {
			if len(path.dangerZones) < len(bestPath.dangerZones) {
				bestPath = path
			}
		}
	}

	return bestPath, true
}
