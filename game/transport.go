package game

import (
	"hermannm.dev/set"
)

func (game *Game) resolveTransports(region *Region) (mustWait bool) {
	type DangerZoneTransport struct {
		move        Order
		dangerZones []DangerZone
	}

	// If a transport is attacked, we must wait to resolve the region, and may end up back here
	// again. Therefore, we don't want to resolve danger zone crossings right away, as they may then
	// get resolved twice. So instead, we store transports through danger zones in this slice, and
	// only resolve them if the incoming move loop doesn't exit on attacked transports.
	var dangerZoneTransports []DangerZoneTransport

	for _, move := range region.incomingMoves {
		if region.hasNeighbor(move.Origin) {
			return false
		}

		canTransport, transportAttacked, dangerZones := game.board.findTransportPath(
			move.Origin,
			move.Destination,
		)
		if !canTransport {
			game.board.retreatMove(move)
			continue
		}
		if transportAttacked {
			return true
		}
		if len(dangerZones) != 0 {
			dangerZoneTransports = append(dangerZoneTransports, DangerZoneTransport{
				move:        move,
				dangerZones: dangerZones,
			})
		}
	}

	if len(dangerZoneTransports) != 0 {
		var crossings []DangerZoneCrossing

		for _, transport := range dangerZoneTransports {
			for _, dangerZone := range transport.dangerZones {
				crossing := game.crossDangerZone(transport.move, dangerZone)
				crossings = append(crossings, crossing)

				if !crossing.Survived {
					game.board.killMove(transport.move)
					break // If we fail a crossing, we don't need to cross any more
				}
			}
		}

		if err := game.messenger.SendDangerZoneCrossings(crossings); err != nil {
			game.log.Error(err)
		}
	}

	region.transportsResolved = true
	return false
}

// Checks if a unit can be transported via ship from the given origin to the given destination.
func (board Board) findTransportPath(
	originName RegionName,
	destinationName RegionName,
) (canTransport bool, transportAttacked bool, dangerZones []DangerZone) {
	origin := board[originName]
	if origin.empty() || origin.Unit.Type == UnitShip || origin.Sea {
		return false, false, nil
	}

	return board.recursivelyFindTransportPath(origin, destinationName, &set.ArraySet[RegionName]{})
}

// Stores status of a path of transport orders to destination.
type transportPath struct {
	attacked    bool
	dangerZones []DangerZone
}

func (board Board) recursivelyFindTransportPath(
	region *Region,
	destination RegionName,
	regionsToExclude set.Set[RegionName],
) (canTransport bool, transportAttacked bool, dangerZones []DangerZone) {
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
					attacked:    transportRegion.attacked(),
					dangerZones: []DangerZone{destinationDangerZone},
				},
			)
		}
		if nextCanTransport {
			subPaths = append(
				subPaths,
				transportPath{
					attacked:    transportRegion.attacked() || nextTransportAttacked,
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
	return canTransport, bestPath.attacked, bestPath.dangerZones
}

func (region *Region) getTransportingNeighbors(
	board Board,
	regionsToExclude set.Set[RegionName],
) (transports []Neighbor, newRegionsToExclude set.Set[RegionName]) {
	newRegionsToExclude = regionsToExclude.Copy()

	if region.empty() {
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
