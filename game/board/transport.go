package board

// Resolves transport of the given move to the given destination if it requires transport.
// If transported, returns whether the transport path is attacked,
// and a list of danger zones that the order must cross to transport, if any.
func (board Board) resolveTransports(move Order, destination Area) (transportAttacked bool, dangerZones []string) {
	adjacent := destination.HasNeighbor(move.From)
	if adjacent {
		return false, nil
	}

	from := board.Areas[move.From]
	if from.Sea {
		return false, nil
	}

	canTransport, transportAttacked, dangerZones := from.canTransportTo(move.To, board, make(map[string]struct{}))

	if !canTransport {
		board.removeMove(move)
		return false, nil
	}

	return transportAttacked, dangerZones
}

// Checks if a unit from the area can be transported to an area with the same name as the given destination.
// Returns whether the unit can be transported, and if so, whether the transports are attacked,
// as well as any potential danger zones the transported unit must cross.
func (area Area) CanTransportTo(destination string, board Board) (
	canTransport bool,
	transportAttacked bool,
	dangerZones []string,
) {
	if area.IsEmpty() || area.Unit.Type == UnitShip {
		return false, false, nil
	}

	return area.canTransportTo(destination, board, make(map[string]struct{}))
}

// Stores status of a path of transport orders to destination.
type transportPath struct {
	attacked    bool
	dangerZones []string
}

// Recursively checks neighbors of the area for available transports to the destination.
// Takes a map of area names to exclude.
func (area Area) canTransportTo(destination string, board Board, exclude map[string]struct{}) (
	canTransport bool,
	transportAttacked bool,
	dangerZones []string,
) {
	transportingNeighbors, newExclude := area.transportingNeighbors(board, exclude)

	// Declares a list of potential transport paths to destination, in order to compare them.
	var paths []transportPath

	// Goes through the area's transporting neighbors to find potential transport paths through them.
	for _, transportNeighbor := range transportingNeighbors {
		transportArea := board.Areas[transportNeighbor.Name]

		attacked := len(transportArea.IncomingMoves) > 0
		destAdjacent, destDangerZone := area.findDestination(destination)

		// Recursively calls this function on the transporting neighbor,
		// in order to find potential transport chains.
		nextCanTransport, nextTransportAttacked, nextDangerZones := transportArea.canTransportTo(
			destination,
			board,
			newExclude,
		)

		var subPaths []transportPath
		if destAdjacent {
			subPaths = append(subPaths, transportPath{
				attacked:    attacked,
				dangerZones: []string{destDangerZone},
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
		if bestPath, canTransport := bestTransportPath(subPaths); canTransport {
			paths = append(paths, bestPath)
		}
	}

	bestPath, canTransport := bestTransportPath(paths)
	return canTransport, bestPath.attacked, bestPath.dangerZones
}

// Finds the given area's friendly neighbors that offer transports.
// Takes a map of area names to exclude,
// and returns a copy of it with the transporting neighbors added.
func (area Area) transportingNeighbors(board Board, exclude map[string]struct{}) (
	transports []Neighbor,
	newExclude map[string]struct{},
) {
	transports = make([]Neighbor, 0)

	newExclude = make(map[string]struct{})
	for excluded := range exclude {
		newExclude[excluded] = struct{}{}
	}

	if area.IsEmpty() {
		return transports, newExclude
	}

	for _, neighbor := range area.Neighbors {
		neighborArea := board.Areas[neighbor.Name]
		_, excluded := exclude[neighbor.Name]

		if excluded ||
			neighborArea.Order.Type != OrderTransport ||
			neighborArea.Unit.Player != area.Unit.Player {
			continue
		}

		transports = append(transports, neighbor)
		newExclude[neighbor.Name] = struct{}{}
	}

	return transports, newExclude
}

// Returns whether the area is adjacent to the given destination,
// and whether a move to it must pass through a danger zone.
func (area Area) findDestination(destination string) (adjacent bool, dangerZone string) {
	for _, neighbor := range area.Neighbors {
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

// From the given paths, returns the best path.
// Prioritizes paths that are not attacked first, then paths that have to cross the fewest danger zones.
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
