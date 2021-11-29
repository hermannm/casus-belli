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

func (area *BoardArea) HasNeighbor(neighborName string) bool {
	for _, neighbor := range area.Neighbors {
		if neighbor.Area.Name == neighborName {
			return true
		}
	}

	return false
}

func (area *BoardArea) NeighborAreas() []*BoardArea {
	areas := make([]*BoardArea, 0)
	added := make(map[string]bool)

	for _, neighbor := range area.Neighbors {
		if added[neighbor.Area.Name] {
			continue
		}

		areas = append(areas, neighbor.Area)
		added[neighbor.Area.Name] = true
	}

	return areas
}
