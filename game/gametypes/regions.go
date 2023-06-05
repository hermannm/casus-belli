package gametypes

// A region on the board.
type Region struct {
	Name      string     `json:"name"`
	Neighbors []Neighbor `json:"neighbors"`

	// Whether the region is a sea region that can only have ship units.
	IsSea bool `json:"isSea"`

	// For land regions: affects the difficulty of conquering the region.
	IsForest bool `json:"isForest,omitempty"`

	// For land regions: affects the difficulty of conquering the region, and the points gained from
	// it.
	HasCastle bool `json:"hasCastle,omitempty"`

	// For land regions: the collection of regions that the region belongs to (affects units gained
	// from conquering).
	Nation string `json:"nation,omitempty"`

	// For land regions that are a starting region for a player.
	HomePlayer string `json:"homePlayer,omitempty"`

	// The unit that currently occupies the region.
	Unit Unit `json:"unit"`

	// The player that currently controls the region.
	ControllingPlayer string `json:"controllingPlayer,omitempty"`

	// For land regions with castles: the number of times an occupying unit has besieged the castle.
	SiegeCount int `json:"siegeCount,omitempty"`

	// Order for the occupying unit in the region. Resets every round.
	Order Order `json:"-"` // Excluded from JSON responses.

	// Incoming move orders to the region. Resets every round.
	IncomingMoves []Order `json:"-"` // Excluded from JSON responses.

	// Incoming support orders to the region. Resets every round.
	IncomingSupports []Order `json:"-"` // Excluded from JSON responses.
}

type Neighbor struct {
	Name string `json:"name"`

	// Whether a river separates the neighboring regions, or this region is a sea and the neighbor
	// is a land region.
	IsAcrossWater bool `json:"isAcrossWater,omitempty"`

	// Whether coast between neighboring land regions have cliffs (impassable to ships).
	HasCliffs bool `json:"hasCliffs,omitempty"`

	// If not "": the name of the danger zone that the neighboring region lies across (requires
	// check to pass).
	DangerZone string `json:"dangerZone,omitempty"`
}

// Checks whether the region contains a unit.
func (region Region) IsEmpty() bool {
	return region.Unit.IsNone()
}

// Checks whether the region is controlled by a player.
func (region Region) IsControlled() bool {
	return region.ControllingPlayer != ""
}

// Checks if any players have move orders against the region.
func (region Region) IsAttacked() bool {
	return len(region.IncomingMoves) != 0
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
	if region.IsSea {
		return false
	}

	for _, neighbor := range region.Neighbors {
		neighborRegion := board.Regions[neighbor.Name]

		if neighborRegion.IsSea {
			return true
		}
	}

	return false
}
