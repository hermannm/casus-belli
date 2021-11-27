package game

func AppendUnitMod(mods []Modifier, unitType UnitType) []Modifier {
	switch unitType {
	case Footman:
		return append(mods, Modifier{
			Type:  UnitMod,
			Value: +1,
		})
	default:
		return mods
	}
}

func DefenseModifiers(area BoardArea) []Modifier {
	mods := []Modifier{}

	mods = AppendUnitMod(mods, area.Unit.Type)

	return mods
}

func AttackModifiers(order Order, otherAttackers bool) []Modifier {
	mods := []Modifier{}

	if (order.To.Control == Uncontrolled && !otherAttackers) ||
		(order.To.Unit != nil && order.To.Control == order.To.Unit.Color) {
		if order.To.Forest {
			mods = append(mods, Modifier{
				Type:  ForestMod,
				Value: -1,
			})
		}
		if order.To.Castle {
			mods = append(mods, Modifier{
				Type:  CastleMod,
				Value: -1,
			})
		}
		if neighbor, ok := order.From.Neighbors[order.To.Name]; ok {
			if neighbor.AcrossWater {
				mods = append(mods, Modifier{
					Type:  WaterMod,
					Value: -1,
				})
			}
		}
	}

	if neighbor, ok := order.From.Neighbors[order.To.Name]; ok {
		if neighbor.DangerZone != "" {
			mods = append(mods, Modifier{
				Type:  SurpriseMod,
				Value: +1,
			})
		}
	}

	if order.From.Unit.Type == Catapult && order.To.Castle {
		mods = append(mods, Modifier{
			Type:  UnitMod,
			Value: +1,
		})
	} else {
		mods = AppendUnitMod(mods, order.From.Unit.Type)
	}

	return mods
}
