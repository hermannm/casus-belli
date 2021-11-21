package types

func (area BoardArea) Sailable() bool {
	if area.Sea {
		return true
	}

	for _, neighbor := range area.Neighbors {
		if neighbor.Area.Sea {
			return true
		}
	}

	return false
}

func (unit Unit) CombatBonus() int {
	switch unit.Type {
	case "footman":
		return 1
	default:
		return 0
	}
}
