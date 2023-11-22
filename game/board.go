package game

import "slices"

type Board map[RegionName]*Region

type RegionName string

// A region on the board.
type Region struct {
	Name      RegionName
	Neighbors []Neighbor

	// Whether the region is a sea region that can only have ship units.
	Sea bool

	// For land regions: affects the difficulty of conquering the region.
	Forest bool

	// For land regions: affects the difficulty of conquering the region, and the points gained from
	// it.
	Castle bool

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

	regionResolvingState
}

// Internal resolving state for a region. Resets to its zero value every round.
// Since all fields are private, they're not included in JSON messages.
type regionResolvingState struct {
	order                    *Order
	incomingMoves            []Order
	incomingSupports         []Order
	expectedSecondHorseMoves int
	incomingSecondHorseMoves []Order
	resolving                bool
	resolved                 bool
	transportsResolved       bool
	unresolvedRetreat        *Order
	partOfCycle              bool // Whether the region is part of a cycle of move orders.
}

type Neighbor struct {
	Name RegionName

	// Whether a river separates the neighboring regions, or this region is a sea and the neighbor
	// is a land region.
	AcrossWater bool

	// Whether coast between neighboring land regions have cliffs (impassable to ships).
	Cliffs bool

	// If not "": the name of the danger zone that the neighboring region lies across (requires
	// check to pass).
	DangerZone DangerZone `json:",omitempty"`
}

type DangerZone string

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
		if !board[supportOrder.Origin].attacked() {
			board.addOrder(supportOrder)
		}
	}
}

func (board Board) addOrder(order Order) {
	origin := board[order.Origin]
	origin.order = &order

	if order.Destination == "" {
		return
	}

	destination := board[order.Destination]
	switch order.Type {
	case OrderMove:
		destination.incomingMoves = append(destination.incomingMoves, order)
		if order.hasSecondHorseMove() {
			destination.expectedSecondHorseMoves++
		}
	case OrderSupport:
		destination.incomingSupports = append(destination.incomingSupports, order)
	}
}

func (board Board) hasUnresolvedRetreats() bool {
	for _, region := range board {
		if region.unresolvedRetreat != nil {
			return true
		}
	}

	return false
}

func (board Board) hasResolvingRegions() bool {
	for _, region := range board {
		if region.resolving {
			return true
		}
	}

	return false
}

func (board Board) resetResolvingState() {
	for _, region := range board {
		region.regionResolvingState = regionResolvingState{}
	}
}

func (board Board) removeOrder(order *Order) {
	if !order.Retreat {
		board[order.Origin].order = nil
	}

	switch order.Type {
	case OrderMove:
		destination := board[order.Destination]
		for i, move := range destination.incomingMoves {
			if move == *order {
				destination.incomingMoves = slices.Delete(destination.incomingMoves, i, i+1)
				break
			}
		}
	case OrderSupport:
		destination := board[order.Destination]
		for i, support := range destination.incomingSupports {
			if support == *order {
				destination.incomingSupports = slices.Delete(destination.incomingSupports, i, i+1)
				break
			}
		}
	}
}

func (board Board) succeedMove(move *Order) {
	destination := board[move.Destination]

	destination.replaceUnit(move.unit)
	destination.order = nil
	if !destination.Sea {
		destination.ControllingFaction = move.Faction
	}

	board[move.Origin].removeUnit()
	board.removeOrder(move)

	destination.resolved = true

	if move.hasSecondHorseMove() {
		secondHorseMove := move.secondHorseMove()
		destination := board[secondHorseMove.Destination]
		destination.incomingSecondHorseMoves = append(
			destination.incomingSecondHorseMoves,
			secondHorseMove,
		)
	}
}

func (board Board) killMove(move *Order) {
	board.removeOrder(move)
	if !move.Retreat {
		board[move.Origin].removeUnit()

		if move.hasSecondHorseMove() {
			board[move.SecondDestination].expectedSecondHorseMoves--
		}
	}
}

func (board Board) retreatMove(move *Order) {
	board.removeOrder(move)

	origin := board[move.Origin]
	if !origin.attacked() {
		origin.Unit = move.unit
	} else if origin.partOfCycle {
		origin.unresolvedRetreat = move
	} else if !move.Retreat {
		retreat := *move
		retreat.Retreat = true
		retreat.Origin, retreat.Destination = retreat.Destination, retreat.Origin
		retreat.SecondDestination = ""
		origin.incomingMoves = append(origin.incomingMoves, retreat)
		origin.order = nil
		origin.removeUnit()
	}

	if move.hasSecondHorseMove() {
		board[move.SecondDestination].expectedSecondHorseMoves--
	}
}

// Checks whether the region contains a unit.
func (region *Region) empty() bool {
	return region.Unit.isNone()
}

// Checks whether the region is controlled by a player faction.
func (region *Region) controlled() bool {
	return region.ControllingFaction != ""
}

// Checks if any players have move orders against the region.
func (region *Region) attacked() bool {
	return len(region.incomingMoves) != 0
}

func (region *Region) removeUnit() {
	// If the region is part of a move cycle, then the unit has already been removed, and another
	// unit may have taken its place
	if !region.partOfCycle {
		region.Unit = Unit{}
		region.SiegeCount = 0
	}
}

func (region *Region) replaceUnit(unit Unit) {
	region.Unit = unit
	region.SiegeCount = 0
}

// Returns a region's neighbor of the given name, and whether it was found.
// If the region has several neighbor relations to the region, returns the one matching the provided
// 'via' string (currently the name of the neighbor relation's danger zone).
func (region *Region) getNeighbor(neighborName RegionName, via DangerZone) (Neighbor, bool) {
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
func (region *Region) hasNeighbor(neighborName RegionName) bool {
	for _, neighbor := range region.Neighbors {
		if neighbor.Name == neighborName {
			return true
		}
	}

	return false
}

// Returns whether the region is a land region that borders the sea.
// Takes the board in order to go through the region's neighbor regions.
func (region *Region) isCoast(board Board) bool {
	if region.Sea {
		return false
	}

	for _, neighbor := range region.Neighbors {
		neighborRegion := board[neighbor.Name]

		if neighborRegion.Sea {
			return true
		}
	}

	return false
}
