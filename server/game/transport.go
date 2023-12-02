package game

import (
	"hermannm.dev/set"
)

func (board Board) resolveUncontestedTransports(region *Region) (mustWait bool) {
	if region.transportsResolved {
		return false
	}

	for _, move := range region.incomingMoves {
		attacked, dangerZone := board.resolveTransport(move, region)
		if attacked || dangerZone != "" {
			return true
		}
	}

	region.transportsResolved = true
	return false
}

func (game *Game) resolveContestedTransports(region *Region) (mustWait bool) {
	if region.transportsResolved {
		return false
	}

	var dangerZoneCrossings []Battle
	for _, move := range region.incomingMoves {
		attacked, dangerZone := game.board.resolveTransport(move, region)
		if attacked {
			return true
		}

		if dangerZone != "" {
			dangerZoneCrossings = append(
				dangerZoneCrossings,
				newDangerZoneCrossing(move, dangerZone),
			)
		}
	}

	for _, crossing := range dangerZoneCrossings {
		game.resolveDangerZoneCrossing(crossing)
	}

	region.transportsResolved = true
	return false
}

func (board Board) resolveTransport(
	move Order,
	destination *Region,
) (transportsAttacked bool, dangerZone DangerZone) {
	if destination.adjacentTo(move.Origin) {
		return false, ""
	}

	canTransport, transportAttacked, dangerZone := board.findTransportPath(
		move.Origin,
		move.Destination,
	)
	if !canTransport {
		board.retreatMove(move)
		return false, ""
	}

	return transportAttacked, dangerZone
}

// Checks if a unit can be transported via ship from the given origin to the given destination.
func (board Board) findTransportPath(
	originName RegionName,
	destinationName RegionName,
) (canTransport bool, transportAttacked bool, dangerZone DangerZone) {
	origin := board[originName]
	if origin.empty() || origin.Unit.Type == UnitShip || origin.Sea {
		return false, false, ""
	}

	return board.recursivelyFindTransportPath(origin, destinationName, &set.ArraySet[RegionName]{})
}

// Stores status of a path of transport orders to destination.
type transportPath struct {
	attacked   bool
	dangerZone DangerZone
}

func (board Board) recursivelyFindTransportPath(
	region *Region,
	destination RegionName,
	excludedRegions set.Set[RegionName],
) (canTransport bool, transportAttacked bool, dangerZone DangerZone) {
	transportingNeighbors, newExcludedRegions := region.getTransportingNeighbors(
		board,
		excludedRegions,
	)

	var paths []transportPath

	for _, transportNeighbor := range transportingNeighbors {
		transportRegion := board[transportNeighbor.Name]

		destinationAdjacent, destinationDangerZone := transportRegion.checkNeighborsForDestination(
			destination,
		)

		// Recursively calls this function on the transporting neighbor,
		// in order to find potential transport chains.
		nextCanTransport, nextTransportAttacked, nextDangerZone := board.recursivelyFindTransportPath(
			transportRegion,
			destination,
			newExcludedRegions,
		)

		var subPaths []transportPath
		if destinationAdjacent {
			subPaths = append(
				subPaths,
				transportPath{
					attacked:   transportRegion.attacked(),
					dangerZone: destinationDangerZone,
				},
			)
		}
		if nextCanTransport {
			subPaths = append(
				subPaths,
				transportPath{
					attacked:   transportRegion.attacked() || nextTransportAttacked,
					dangerZone: nextDangerZone,
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
	return canTransport, bestPath.attacked, bestPath.dangerZone
}

func (region *Region) getTransportingNeighbors(
	board Board,
	excludedRegions set.Set[RegionName],
) (transports []Neighbor, newExcludedRegions set.Set[RegionName]) {
	newExcludedRegions = excludedRegions.Copy()

	if region.empty() {
		return transports, newExcludedRegions
	}

	for _, neighbor := range region.Neighbors {
		neighborRegion := board[neighbor.Name]

		if excludedRegions.Contains(neighbor.Name) ||
			neighborRegion.order.Type != OrderTransport ||
			neighborRegion.Unit.Faction != region.Unit.Faction {
			continue
		}

		transports = append(transports, neighbor)
		newExcludedRegions.Add(neighbor.Name)
	}

	return transports, newExcludedRegions
}

func (region *Region) checkNeighborsForDestination(
	destination RegionName,
) (destinationAdjacent bool, mustGoThrough DangerZone) {
	for _, neighbor := range region.Neighbors {
		if neighbor.Name == destination {
			// If destination is already found to be adjacent but only through a danger zone,
			// checks if there is a different path to it without a danger zone.
			if destinationAdjacent {
				if mustGoThrough != "" {
					mustGoThrough = neighbor.DangerZone
				}
				continue
			}

			destinationAdjacent = true
			mustGoThrough = neighbor.DangerZone
		}
	}

	return destinationAdjacent, mustGoThrough
}

// Prioritizes paths that are not attacked first, then paths that don't have to cross danger zones.
func bestTransportPath(paths []transportPath) (bestPath transportPath, canTransport bool) {
	if len(paths) == 0 {
		return transportPath{}, false
	}

	bestPath = paths[0]

	for _, path := range paths[1:] {
		if bestPath.attacked {
			if path.attacked {
				if path.dangerZone == "" && bestPath.dangerZone != "" {
					bestPath = path
				}
			} else {
				bestPath = path
			}
		} else if !path.attacked {
			if path.dangerZone == "" && bestPath.dangerZone != "" {
				bestPath = path
			}
		}
	}

	return bestPath, true
}
