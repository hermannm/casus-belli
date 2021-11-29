package game

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

func (order Order) Transportable() (
	transportable bool,
	dangerZone bool,
) {
	transportable, dangerZone = order.From.canNeighborsTransport(order.To.Name, make(map[string]bool))

	return transportable, dangerZone
}

func (area BoardArea) transportingNeighbors(exclude map[string]bool) (
	neighbors []Neighbor,
	newExclude map[string]bool,
) {
	neighbors = make([]Neighbor, 0)
	newExclude = make(map[string]bool)
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

func (area BoardArea) canNeighborsTransport(destination string, exclude map[string]bool) (
	transportable bool,
	dangerZone bool,
) {
	dangerZone = true

	transportingNeighbors, newExclude := area.transportingNeighbors(exclude)

	for _, transport := range transportingNeighbors {
		if len(transport.Area.IncomingMoves) > 0 {
			continue
		}

		if transport.Area.HasNeighbor(destination) {
			transportable = true

			if transport.DangerZone == "" {
				dangerZone = false
			}
		} else {
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
