package gametypes

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

	return board.recursivelyFindTransportPath(origin, destinationName, make(map[string]struct{}))
}

// Stores status of a path of transport orders to destination.
type transportPath struct {
	attacked    bool
	dangerZones []string
}

// Recursively checks neighbors of the region for available transports to the destination.
// Takes a map of region names to exclude.
func (board Board) recursivelyFindTransportPath(
	region Region, destination string, exclude map[string]struct{},
) (canTransport bool, transportAttacked bool, dangerZones []string) {
	transportingNeighbors, newExclude := region.getTransportingNeighbors(board, exclude)

	// Declares a list of potential transport paths to destination, in order to compare them.
	var paths []transportPath

	// Goes through the region's transporting neighbors to find potential transport paths.
	for _, transportNeighbor := range transportingNeighbors {
		transportRegion := board.Regions[transportNeighbor.Name]

		attacked := len(transportRegion.IncomingMoves) > 0
		destinationAdjacent, destinationDangerZone := region.
			checkNeighborsForDestination(destination)

		// Recursively calls this function on the transporting neighbor,
		// in order to find potential transport chains.
		nextCanTransport, nextTransportAttacked, nextDangerZones := board.
			recursivelyFindTransportPath(transportRegion, destination, newExclude)

		var subPaths []transportPath
		if destinationAdjacent {
			subPaths = append(subPaths, transportPath{
				attacked:    attacked,
				dangerZones: []string{destinationDangerZone},
			})
		}
		if nextCanTransport {
			subPaths = append(subPaths, transportPath{
				attacked:    attacked || nextTransportAttacked,
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
	board Board, exclude map[string]struct{},
) (transports []Neighbor, newExclude map[string]struct{}) {
	transports = make([]Neighbor, 0)

	newExclude = make(map[string]struct{})
	for excluded := range exclude {
		newExclude[excluded] = struct{}{}
	}

	if region.IsEmpty() {
		return transports, newExclude
	}

	for _, neighbor := range region.Neighbors {
		neighborRegion := board.Regions[neighbor.Name]
		_, excluded := exclude[neighbor.Name]

		if excluded ||
			neighborRegion.Order.Type != OrderTransport ||
			neighborRegion.Unit.Player != region.Unit.Player {
			continue
		}

		transports = append(transports, neighbor)
		newExclude[neighbor.Name] = struct{}{}
	}

	return transports, newExclude
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
