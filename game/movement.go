package game

func (area *BoardArea) GetNeighbor(neighborName string, via string) (Neighbor, bool) {
	var n Neighbor
	ok := false

	for _, neighbor := range area.Neighbors {
		if neighborName == neighbor.Area.Name {
			if !ok {
				n = neighbor
				ok = true
			} else if neighbor.DangerZone != "" && via == neighbor.DangerZone {
				n = neighbor
			}
		}
	}

	return n, ok
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

func (area *BoardArea) TransportNeighbors(path []*BoardArea) (
	neighbors map[string]*BoardArea,
	paths [][]*BoardArea,
) {
	neighbors = make(map[string]*BoardArea)
	paths = make([][]*BoardArea, 0)

outer:
	for _, neighborArea := range area.NeighborAreas() {
		if neighborArea.Sea {
			for _, alreadyTransporting := range path {
				if alreadyTransporting.Name == neighborArea.Name {
					continue outer
				}
			}

			if neighborArea.Outgoing != nil &&
				neighborArea.Outgoing.Type == Transport &&
				neighborArea.Unit.Color == area.Unit.Color {

				connectedNeighbors, paths := neighborArea.TransportNeighbors(append(path, area))

				for _, newPath := range paths {
					paths = append(paths, append(path, newPath...))
				}

				neighbors = mergeMaps(neighbors, connectedNeighbors)
			}
		} else if area.Sea {
			neighbors[neighborArea.Name] = neighborArea
		}
	}

	return neighbors, paths
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
