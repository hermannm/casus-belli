package game

type Board map[RegionName]Region

type RegionName string

// A region on the board.
type Region struct {
	Name      RegionName
	Neighbors []Neighbor

	// Whether the region is a sea region that can only have ship units.
	IsSea bool

	// For land regions: affects the difficulty of conquering the region.
	IsForest bool

	// For land regions: affects the difficulty of conquering the region, and the points gained from
	// it.
	HasCastle bool

	// For land regions: the collection of regions that the region belongs to (affects units gained
	// from conquering).
	Nation string `json:",omitempty"`

	// For land regions that are a starting region for a player faction.
	HomeFaction PlayerFaction `json:",omitempty"`

	// The unit that currently occupies the region.
	Unit Unit

	// The player faction that currently controls the region.
	ControllingFaction PlayerFaction `json:",omitempty"`

	// For land regions with castles: the number of times an occupying unit has besieged the castle.
	SiegeCount int `json:",omitempty"`

	// Order for the occupying unit in the region. Resets every round.
	Order Order `json:"-"` // Excluded from JSON responses.

	// Incoming move orders to the region. Resets every round.
	IncomingMoves []Order `json:"-"` // Excluded from JSON responses.

	// Incoming support orders to the region. Resets every round.
	IncomingSupports []Order `json:"-"` // Excluded from JSON responses.
}

type Neighbor struct {
	Name RegionName

	// Whether a river separates the neighboring regions, or this region is a sea and the neighbor
	// is a land region.
	IsAcrossWater bool

	// Whether coast between neighboring land regions have cliffs (impassable to ships).
	HasCliffs bool

	// If not "": the name of the danger zone that the neighboring region lies across (requires
	// check to pass).
	DangerZone DangerZone `json:",omitempty"`
}

type DangerZone string

func (board Board) removeUnit(unit Unit, regionName RegionName) {
	region, ok := board[regionName]
	if !ok {
		return
	}

	if unit == region.Unit {
		region.Unit = Unit{}
		board[regionName] = region
	}
}

// Populates regions on the board with the given orders.
// Does not add support orders that have moves against them, as that cancels them.
func (board Board) addOrders(orders []Order) {
	var supportOrders []Order

	for _, order := range orders {
		if order.Type == OrderSupport {
			supportOrders = append(supportOrders, order)
			continue
		}

		board.addOrder(order)
	}

	for _, supportOrder := range supportOrders {
		if !board[supportOrder.Origin].isAttacked() {
			board.addOrder(supportOrder)
		}
	}
}

func (board Board) addOrder(order Order) {
	origin := board[order.Origin]
	origin.Order = order
	board[order.Origin] = origin

	if order.Destination == "" {
		return
	}

	destination := board[order.Destination]
	switch order.Type {
	case OrderMove:
		destination.IncomingMoves = append(destination.IncomingMoves, order)
	case OrderSupport:
		destination.IncomingSupports = append(destination.IncomingSupports, order)
	}
	board[order.Destination] = destination
}

func (board Board) clearOrders() {
	for regionName, region := range board {
		region.Order = Order{}
		region.IncomingMoves = nil
		region.IncomingSupports = nil

		board[regionName] = region
	}
}

func (board Board) removeOrder(order Order) {
	origin := board[order.Origin]
	origin.Order = Order{}
	board[order.Origin] = origin

	switch order.Type {
	case OrderMove:
		destination := board[order.Destination]

		var newMoves []Order
		for _, incomingMove := range destination.IncomingMoves {
			if incomingMove != order {
				newMoves = append(newMoves, incomingMove)
			}
		}
		destination.IncomingMoves = newMoves

		board[order.Destination] = destination
	case OrderSupport:
		destination := board[order.Destination]

		var newSupports []Order
		for _, incSupport := range destination.IncomingSupports {
			if incSupport != order {
				newSupports = append(newSupports, incSupport)
			}
		}
		destination.IncomingSupports = newSupports

		board[order.Destination] = destination
	}
}

// Checks whether the region contains a unit.
func (region Region) isEmpty() bool {
	return region.Unit.isNone()
}

// Checks whether the region is controlled by a player faction.
func (region Region) isControlled() bool {
	return region.ControllingFaction != ""
}

// Checks if any players have move orders against the region.
func (region Region) isAttacked() bool {
	return len(region.IncomingMoves) != 0
}

// Returns a region's neighbor of the given name, and whether it was found.
// If the region has several neighbor relations to the region, returns the one matching the provided
// 'via' string (currently the name of the neighbor relation's danger zone).
func (region Region) getNeighbor(neighborName RegionName, via DangerZone) (Neighbor, bool) {
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
func (region Region) hasNeighbor(neighborName RegionName) bool {
	for _, neighbor := range region.Neighbors {
		if neighbor.Name == neighborName {
			return true
		}
	}

	return false
}

// Returns whether the region is a land region that borders the sea.
// Takes the board in order to go through the region's neighbor regions.
func (region Region) isCoast(board Board) bool {
	if region.IsSea {
		return false
	}

	for _, neighbor := range region.Neighbors {
		neighborRegion := board[neighbor.Name]

		if neighborRegion.IsSea {
			return true
		}
	}

	return false
}
