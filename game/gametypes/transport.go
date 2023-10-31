package gametypes

import (
	"hermannm.dev/set"
)

// Checks if a unit can be transported via ship from the given origin to the given destination.
func (board Board) FindTransportPath(
	originName string,
	destinationName string,
) (canTransport bool, isTransportAttacked bool, dangerZones []string) {
	origin := board.Regions[originName]
	if origin.IsEmpty() || origin.Unit.Type == UnitShip || origin.IsSea {
		return false, false, nil
	}

	return board.recursivelyFindTransportPath(origin, destinationName, &set.ArraySet[string]{})
}

// Stores status of a path of transport orders to destination.
type transportPath struct {
	isAttacked  bool
	dangerZones []string
}

func (board Board) recursivelyFindTransportPath(
	region Region,
	destination string,
	regionsToExclude set.Set[string],
) (canTransport bool, isTransportAttacked bool, dangerZones []string) {
	transportingNeighbors, newRegionsToExclude := region.getTransportingNeighbors(
		board,
		regionsToExclude,
	)

	var paths []transportPath

	for _, transportNeighbor := range transportingNeighbors {
		transportRegion := board.Regions[transportNeighbor.Name]

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
					isAttacked:  transportRegion.IsAttacked(),
					dangerZones: []string{destinationDangerZone},
				},
			)
		}
		if nextCanTransport {
			subPaths = append(
				subPaths,
				transportPath{
					isAttacked:  transportRegion.IsAttacked() || nextTransportAttacked,
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

func (region Region) getTransportingNeighbors(
	board Board,
	regionsToExclude set.Set[string],
) (transports []Neighbor, newRegionsToExclude set.Set[string]) {
	newRegionsToExclude = regionsToExclude.Copy()

	if region.IsEmpty() {
		return transports, newRegionsToExclude
	}

	for _, neighbor := range region.Neighbors {
		neighborRegion := board.Regions[neighbor.Name]

		if regionsToExclude.Contains(neighbor.Name) ||
			neighborRegion.Order.Type != OrderTransport ||
			neighborRegion.Unit.Player != region.Unit.Player {
			continue
		}

		transports = append(transports, neighbor)
		newRegionsToExclude.Add(neighbor.Name)
	}

	return transports, newRegionsToExclude
}

func (region Region) checkNeighborsForDestination(
	destination string,
) (destinationIsAdjacent bool, throughDangerZone string) {
	for _, neighbor := range region.Neighbors {
		if neighbor.Name == destination {
			// If destination is already found to be adjacent but only through a danger zone,
			// checks if there is a different path to it without a danger zone.
			if destinationIsAdjacent {
				if throughDangerZone != "" {
					throughDangerZone = neighbor.DangerZone
				}
				continue
			}

			destinationIsAdjacent = true
			throughDangerZone = neighbor.DangerZone
		}
	}

	return destinationIsAdjacent, throughDangerZone
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
