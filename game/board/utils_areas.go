package board

// Checks whether the area contains a unit.
func (area Area) IsEmpty() bool {
	return area.Unit.Type == ""
}

// Checks whether the area is controlled by a player.
func (area Area) IsControlled() bool {
	return area.Control != ""
}

// Returns an area's neighbor of the given name, and whether it was found.
// If the area has several neighbor relations to the area, returns the one matching the provided 'via' string
// (currently the name of the neighbor relation's danger zone).
func (area Area) GetNeighbor(neighborName string, via string) (Neighbor, bool) {
	neighbor := Neighbor{}
	hasNeighbor := false

	for _, otherNeighbor := range area.Neighbors {
		if neighborName != otherNeighbor.Name {
			continue
		}

		if !hasNeighbor {
			neighbor = otherNeighbor
			hasNeighbor = true
		} else if otherNeighbor.DangerZone != "" && via == otherNeighbor.DangerZone {
			neighbor = otherNeighbor
		}
	}

	return neighbor, hasNeighbor
}

// Returns whether the area is adjacent to an area of the given name.
func (area Area) HasNeighbor(neighborName string) bool {
	for _, neighbor := range area.Neighbors {
		if neighbor.Name == neighborName {
			return true
		}
	}

	return false
}

// Returns whether the area is a land area that borders the sea.
// Takes the board in order to go through the area's neighbor areas.
func (area Area) IsCoast(board Board) bool {
	if area.Sea {
		return false
	}

	for _, neighbor := range area.Neighbors {
		neighborArea := board.Areas[neighbor.Name]

		if neighborArea.Sea {
			return true
		}
	}

	return false
}

// Returns a copy of the area, with its unit set to the given unit.
func (area Area) setUnit(unit Unit) Area {
	area.Unit = unit
	return area
}

// Returns a copy of the area, with control set to the given player.
func (area Area) setControl(player Player) Area {
	area.Control = player
	return area
}

// Returns a copy of the area, with its order field set to the given order.
func (area Area) setOrder(order Order) Area {
	area.Order = order
	return area
}
