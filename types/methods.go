package types

func (area BoardArea) IsCoast() bool {
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
