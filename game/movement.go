package game

func (area *BoardArea) GetNeighbor(neighborName string, via string) (Neighbor, bool) {
	for _, neighbor := range area.Neighbors {
		if neighborName == neighbor.Area.Name && via == neighbor.DangerZone {
			return neighbor, true
		}
	}

	return Neighbor{}, false
}

func (area *BoardArea) NeighborAreas() []*BoardArea {
	areas := make([]*BoardArea, 0)

	for _, neighbor := range area.Neighbors {
		included := false

		for _, added := range areas {
			if added.Name == neighbor.Area.Name {
				included = true
				break
			}
		}

		if !included {
			areas = append(areas, neighbor.Area)
		}
	}

	return areas
}

func (area *BoardArea) HasNeighbor(neighborName string) bool {
	for _, neighbor := range area.Neighbors {
		if neighbor.Area.Name == neighborName {
			return true
		}
	}

	return false
}

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

	for _, neighborArea := range area.NeighborAreas() {
		if neighborArea.Sea {
			if _, excluded := exclude[neighborArea.Name]; excluded {
				continue
			}

			if neighborArea.Outgoing != nil &&
				neighborArea.Outgoing.Type == Transport &&
				neighborArea.Unit.Color == area.Unit.Color {

				newExclude := copyMap(exclude)
				newExclude[area.Name] = area
				connectedNeighbors := neighborArea.transportNeighbors(newExclude)

				neighbors = mergeMaps(neighbors, connectedNeighbors)
			}
		} else if area.Sea {
			neighbors[neighborArea.Name] = neighborArea
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
