package gametypes

// A region on the board map.
type Region struct {
	// Name of the region on the board.
	Name string `json:"name"`

	// Adjacent regions.
	Neighbors []Neighbor `json:"neighbors"`

	// Whether the region is a sea region that can only have ship units.
	Sea bool `json:"sea"`

	// For land regions: affects the difficulty of conquering the region.
	Forest bool `json:"forest,omitempty"`

	// For land regions: affects the difficulty of conquering the region, and the points gained from
	// it.
	Castle bool `json:"castle,omitempty"`

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

// The relationship between two adjacent regions.
type Neighbor struct {
	// Name of the adjacent region.
	Name string `json:"name"`

	// Whether a river separates the two regions.
	AcrossWater bool `json:"acrossWater,omitempty"`

	// Whether coast between neighboring land regions have cliffs (impassable to ships).
	Cliffs bool `json:"cliffs,omitempty"`

	// If not "": the name of the danger zone that the neighboring region lies across (requires
	// check to pass).
	DangerZone string `json:"dangerZone,omitempty"`
}

// Checks whether the region contains a unit.
func (region Region) IsEmpty() bool {
	return region.Unit.Type == ""
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
