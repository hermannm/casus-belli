package game

func (area *BoardArea) IsCoast() bool {
	if area.Sea {
		return false
	}

	for _, neighbor := range area.Neighbors {
		if neighbor.Area.Sea {
			return true
		}
	}

	return false
}

func (order *Order) Transportable() bool {
	if order.Type != Move || order.From.Unit.Type == Ship {
		return false
	}

	possibleDestinations := order.From.transportNeighbors(make(map[string]*BoardArea))

	_, transportable := possibleDestinations[order.To.Name]

	return transportable
}

func (area *BoardArea) transportNeighbors(exclude map[string]*BoardArea) map[string]*BoardArea {
	neighbors := make(map[string]*BoardArea)

	for _, neighbor := range area.Neighbors {
		if neighbor.Area.Sea {
			if _, excluded := exclude[neighbor.Area.Name]; excluded {
				continue
			}

			if neighbor.Area.Outgoing != nil &&
				neighbor.Area.Outgoing.Type == Transport &&
				neighbor.Area.Unit.Color == area.Unit.Color {

				newExclude := copyMap(exclude)
				newExclude[area.Name] = area
				connectedNeighbors := neighbor.Area.transportNeighbors(newExclude)

				neighbors = mergeMaps(neighbors, connectedNeighbors)
			}
		} else if area.Sea {
			neighbors[neighbor.Area.Name] = neighbor.Area
		}
	}

	return neighbors
}

func copyMap(oldMap map[string]*BoardArea) map[string]*BoardArea {
	newMap := make(map[string]*BoardArea)
	for key, area := range oldMap {
		newMap[key] = area
	}
	return newMap
}

func mergeMaps(maps ...map[string]*BoardArea) map[string]*BoardArea {
	newMap := make(map[string]*BoardArea)

	for _, subMap := range maps {
		for key, area := range subMap {
			newMap[key] = area
		}
	}

	return newMap
}
