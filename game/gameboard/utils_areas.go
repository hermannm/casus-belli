package gameboard

// Checks whether the region contains a unit.
func (region Region) IsEmpty() bool {
	return region.Unit.Type == ""
}

// Checks whether the region is controlled by a player.
func (region Region) IsControlled() bool {
	return region.ControllingPlayer != ""
}

// Returns a region's neighbor of the given name, and whether it was found.
// If the region has several neighbor relations to the region, returns the one matching the provided
// 'via' string (currently the name of the neighbor relation's danger zone).
func (region Region) GetNeighbor(neighborName string, via string) (Neighbor, bool) {
	neighbor := Neighbor{}
	hasNeighbor := false

	for _, otherNeighbor := range region.Neighbors {
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

// Returns whether the region is adjacent to a region of the given name.
func (region Region) HasNeighbor(neighborName string) bool {
	for _, neighbor := range region.Neighbors {
		if neighbor.Name == neighborName {
			return true
		}
	}

	return false
}

// Returns whether the region is a land region that borders the sea.
// Takes the board in order to go through the region's neighbor regions.
func (region Region) IsCoast(board Board) bool {
	if region.Sea {
		return false
	}

	for _, neighbor := range region.Neighbors {
		neighborRegion := board.Regions[neighbor.Name]

		if neighborRegion.Sea {
			return true
		}
	}

	return false
}
