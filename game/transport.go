package game

// Fails transport-dependent move if it cannot transport.
// Returns true if the transport did not fail.
func (order *Order) Transport() bool {
	transportable, dangerZone := order.Transportable()

	if transportable {
		if dangerZone {
			return order.crossDangerZone()
		}

		return true
	} else {
		order.failMove()

		return false
	}
}

// Checks if a transport-dependent move can be transported.
// If transportable, also returns whether transport must pass danger zone.
func (order Order) Transportable() (
	transportable bool,
	dangerZone bool,
) {
	transportable, dangerZone = order.From.canNeighborsTransport(order.To.Name, make(map[string]bool))

	return transportable, dangerZone
}

// Checks if a land unit can be transported to destination.
// Takes a map of area names to exclude, to enable recursion.
// Returns whether the unit can be transported, and if so, whether it must pass through danger zone.
func (area BoardArea) canNeighborsTransport(destination string, exclude map[string]bool) (
	transportable bool,
	dangerZone bool,
) {
	dangerZone = true

	transportingNeighbors, newExclude := area.transportingNeighbors(exclude)

	for _, transport := range transportingNeighbors {
		// Transports either happen when resolving conflict-free orders, in which case it should not allow transports under attack,
		// or after transport combats are resolved, in which case there should no longer be any transports under attack.
		if len(transport.Area.IncomingMoves) > 0 {
			continue
		}

		if transport.Area.HasNeighbor(destination) {
			transportable = true

			if transport.DangerZone == "" {
				dangerZone = false
			}
		} else {
			// Recursive call to check for eligible chain of transports to destination.
			canTransport, danger := transport.Area.canNeighborsTransport(destination, newExclude)

			if canTransport {
				transportable = true

				if !danger && transport.DangerZone == "" {
					dangerZone = false
				}
			}
		}
	}

	return transportable, dangerZone
}

// Finds an area's friendly neighbors that offer transports.
// Takes a map of area names to exclude, and returns it with the transporting neighbors added.
func (area BoardArea) transportingNeighbors(exclude map[string]bool) ([]Neighbor, map[string]bool) {
	neighbors := make([]Neighbor, 0)
	newExclude := make(map[string]bool)
	for k, v := range exclude {
		newExclude[k] = v
	}

	if area.Unit == nil {
		return neighbors, newExclude
	}

	for _, neighbor := range area.Neighbors {
		if neighbor.Area.Outgoing != nil &&
			neighbor.Area.Outgoing.Type == Transport &&
			neighbor.Area.Unit.Color == area.Unit.Color {

			if exclude[neighbor.Area.Name] {
				continue
			}

			neighbors = append(neighbors, neighbor)
			newExclude[neighbor.Area.Name] = true
		}
	}

	return neighbors, newExclude
}
