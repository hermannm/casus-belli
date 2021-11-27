package game

func Transportable(order *Order) bool {
	if order.Type != Move || order.From.Unit.Type == Ship {
		return false
	}

	possibleDestinations := findTransportNeighbors(order.From, make(map[string]*BoardArea))

	_, transportable := possibleDestinations[order.To.Name]

	return transportable
}

func findTransportNeighbors(area *BoardArea, exclude map[string]*BoardArea) map[string]*BoardArea {
	neighbors := make(map[string]*BoardArea)

	for _, neighbor := range area.Neighbors {
		if neighbor.Area.Sea {
			if _, excluded := exclude[neighbor.Area.Name]; excluded {
				continue
			}

			if neighbor.Area.Outgoing.Type == Transport && neighbor.Area.Unit.Color == area.Unit.Color {
				newExclude := copyMap(exclude)
				newExclude[area.Name] = area
				connectedNeighbors := findTransportNeighbors(neighbor.Area, newExclude)

				neighbors = mergeMaps(neighbors, connectedNeighbors)
			}
		} else if area.Sea {
			neighbors[neighbor.Area.Name] = neighbor.Area
		}
	}

	return neighbors
}
