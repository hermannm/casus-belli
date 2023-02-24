package gametypes

import (
	"hermannm.dev/set"
)

// Checks if a unit can be transported from the given origin region to the given destination.
// Returns whether the unit can be transported, and if so, whether the transports are attacked,
// as well as any potential danger zones the transported unit must cross.
func (board Board) FindTransportPath(
	originName string, destinationName string,
) (canTransport bool, transportAttacked bool, dangerZones []string) {
	origin := board.Regions[originName]
	if origin.IsEmpty() || origin.Unit.Type == UnitShip || origin.Sea {
		return false, false, nil
	}

	return board.recursivelyFindTransportPath(origin, destinationName, set.New[string]())
}

// Stores status of a path of transport orders to destination.
type transportPath struct {
	attacked    bool
	dangerZones []string
}

// Recursively checks neighbors of the region for available transports to the destination.
// Takes a map of region names to exclude.
func (board Board) recursivelyFindTransportPath(
	region Region, destination string, regionsToExclude set.Set[string],
) (canTransport bool, transportAttacked bool, dangerZones []string) {
	transportingNeighbors, newRegionsToExclude := region.getTransportingNeighbors(
		board, regionsToExclude,
	)

	// Declares a list of potential transport paths to destination, in order to compare them.
	var paths []transportPath

	// Goes through the region's transporting neighbors to find potential transport paths.
	for _, transportNeighbor := range transportingNeighbors {
		transportRegion := board.Regions[transportNeighbor.Name]

		destinationAdjacent, destinationDangerZone := region.
			checkNeighborsForDestination(destination)

		// Recursively calls this function on the transporting neighbor,
		// in order to find potential transport chains.
		nextCanTransport, nextTransportAttacked, nextDangerZones := board.
			recursivelyFindTransportPath(transportRegion, destination, newRegionsToExclude)

		var subPaths []transportPath
		if destinationAdjacent {
			subPaths = append(subPaths, transportPath{
				attacked:    transportRegion.IsAttacked(),
				dangerZones: []string{destinationDangerZone},
			})
		}
		if nextCanTransport {
			subPaths = append(subPaths, transportPath{
				attacked:    transportRegion.IsAttacked() || nextTransportAttacked,
				dangerZones: nextDangerZones,
			})
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
	return canTransport, bestPath.attacked, bestPath.dangerZones
}

// Finds the given region's friendly neighbors that offer transports.
// Takes a map of region names to exclude,
// and returns a copy of it with the transporting neighbors added.
func (region Region) getTransportingNeighbors(
	board Board, regionsToExclude set.Set[string],
) (transports []Neighbor, newRegionsToExclude set.Set[string]) {
	newRegionsToExclude = set.New[string]()
	for excluded := range regionsToExclude {
		newRegionsToExclude.Add(excluded)
	}

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

// Returns whether the region is adjacent to the given destination,
// and whether a move to it must pass through a danger zone.
func (region Region) checkNeighborsForDestination(
	destination string,
) (adjacent bool, dangerZone string) {
	for _, neighbor := range region.Neighbors {
		if neighbor.Name == destination {
			// If destination is already found to be adjacent but only through a danger zone,
			// checks if there is a different path to it without a danger zone.
			if adjacent {
				if dangerZone != "" {
					dangerZone = neighbor.DangerZone
				}
				continue
			}

			adjacent = true
			dangerZone = neighbor.DangerZone
		}
	}

	return adjacent, dangerZone
}

// From the given transport paths, returns the best path. Prioritizes paths that are not attacked
// first, then paths that have to cross the fewest danger zones.
//
// If the given path list contains no paths, returns canTransport = false.
func bestTransportPath(paths []transportPath) (bestPath transportPath, canTransport bool) {
	if len(paths) == 0 {
		return transportPath{}, false
	}

	bestPath = paths[0]

	for _, path := range paths[1:] {
		if bestPath.attacked {
			if path.attacked {
				if len(path.dangerZones) < len(bestPath.dangerZones) {
					bestPath = path
				}
			} else {
				bestPath = path
			}
		} else if !path.attacked {
			if len(path.dangerZones) < len(bestPath.dangerZones) {
				bestPath = path
			}
		}
	}

	return bestPath, true
}
