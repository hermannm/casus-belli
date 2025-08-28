package game

import (
	"slices"

	"hermannm.dev/opt"
	"hermannm.dev/set"
)

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

	// The unit that currently occupies the region (if any).
	Unit opt.Option[Unit]

	// The player faction that currently controls the region.
	ControllingFaction PlayerFaction `json:",omitempty"`

	// For land regions with castles: the number of times an occupying unit has besieged the castle.
	SiegeCount int `json:",omitempty"`

	regionResolvingState
}

// Internal resolving state for a region. Resets to its zero value every round.
// Since all fields are private, they're not included in JSON messages.
type regionResolvingState struct {
	order                opt.Option[Order]
	incomingMoves        []Order
	incomingSupports     []Order
	incomingKnightMoves  []Order
	expectedKnightMoves  int
	resolvingKnightMoves bool
	resolved             bool
	transportsResolved   bool
	dangerZonesResolved  bool
	partOfCycle          bool // Whether the region is part of a cycle of move orders.
	unresolvedRetreat    opt.Option[Order]
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

// Populates regions on the board with the given orders.
// Does not add support orders that have moves against them, as that cancels them.
func (board Board) placeOrders(orders []Order) {
	var supportOrders []Order

	for _, order := range orders {
		if order.Type == OrderSupport {
			supportOrders = append(supportOrders, order)
			continue
		}

		board.placeOrder(order)
	}

	for _, supportOrder := range supportOrders {
		if !board[supportOrder.Origin].attacked() {
			board.placeOrder(supportOrder)
		}
	}
}

func (board Board) placeOrder(order Order) {
	origin := board[order.Origin]
	origin.order.Put(order)

	if order.Destination == "" {
		return
	}

	destination := board[order.Destination]
	switch order.Type {
	case OrderMove:
		destination.incomingMoves = append(destination.incomingMoves, order)
		if order.hasKnightMove() {
			board[order.SecondDestination].expectedKnightMoves++
		}
	case OrderSupport:
		destination.incomingSupports = append(destination.incomingSupports, order)
	default: // Done
	}
}

func (board Board) placeKnightMoves(region *Region) {
	region.incomingMoves = region.incomingKnightMoves
	for _, move := range region.incomingMoves {
		board[move.Origin].order.Put(move)
	}

	region.resolvingKnightMoves = true
	region.incomingKnightMoves = nil
	region.expectedKnightMoves = 0
	region.transportsResolved = false
	region.dangerZonesResolved = false
	region.partOfCycle = false
}

// Assumes that the given region is under attack from knight moves. Cuts any outgoing support order
// from the region, unless we must wait for the support's destination to reach knight move
// resolving.
//
// Also goes through incoming supports to the region, and cuts any incoming support orders that are
// under attack, unless we must wait for their origin regions to reach knight move resolving.
func (board Board) cutSupportsAttackedByKnightMoves(region *Region) (mustWait bool) {
	if order, ok := region.order.Get(); ok && order.Type == OrderSupport {
		destination := board[order.Destination]
		if !destination.resolved && !destination.resolvingKnightMoves {
			return true
		}

		board.removeOrder(order)
	}

	var supportsToCut []Order
	for _, support := range region.incomingSupports {
		origin := board[support.Origin]
		if origin.resolved {
			continue
		}

		if !origin.resolvingKnightMoves {
			mustWait = true
			continue
		}

		if origin.attacked() {
			// Cuts the supports in a loop below, since cutting it here mutates
			// region.incomingSupports, invalidating our loop
			supportsToCut = append(supportsToCut, support)
		}
	}

	for _, support := range supportsToCut {
		board.removeOrder(support)
	}

	return mustWait
}

func (board Board) resolved() bool {
	for _, region := range board {
		if !region.resolved {
			return false
		}
	}
	return true
}

func (board Board) resetResolvingState() {
	for _, region := range board {
		region.regionResolvingState = regionResolvingState{} //nolint:exhaustruct
	}
}

func (board Board) removeOrder(order Order) {
	if !order.Retreat {
		board[order.Origin].order.Clear()
	}

	if order.Type == OrderMove {
		destination := board[order.Destination]
		for i, move := range destination.incomingMoves {
			if move == order {
				destination.incomingMoves = slices.Delete(destination.incomingMoves, i, i+1)
				break
			}
		}
	} else if order.Type == OrderSupport {
		destination := board[order.Destination]
		for i, support := range destination.incomingSupports {
			if support == order {
				destination.incomingSupports = slices.Delete(destination.incomingSupports, i, i+1)
				break
			}
		}
	}
}

func (board Board) succeedMove(move Order) {
	destination := board[move.Destination]

	destination.replaceUnit(move.unit())
	destination.order.Clear()

	// Seas cannot be controlled, and unconquered castles must be besieged first, unless the
	// attacking unit is a catapult
	if !destination.Sea &&
		(!destination.Castle || destination.controlled() || move.UnitType == UnitCatapult) {
		destination.ControllingFaction = move.Faction
	}

	board[move.Origin].removeUnit()
	board.removeOrder(move)

	if move.hasKnightMove() {
		knightMove := move.knightMove()
		destination := board[knightMove.Destination]
		destination.incomingKnightMoves = append(destination.incomingKnightMoves, knightMove)
	}
}

func (board Board) killMove(move Order) {
	board.removeOrder(move)

	if !move.Retreat {
		board[move.Origin].removeUnit()

		if move.hasKnightMove() {
			board[move.SecondDestination].expectedKnightMoves--
		}
	}
}

func (board Board) retreatMove(move Order) {
	board.removeOrder(move)

	if !move.Retreat {
		origin := board[move.Origin]
		if !origin.attacked() {
			origin.Unit.Put(move.unit())
		} else if origin.partOfCycle {
			origin.unresolvedRetreat.Put(move)
		} else if !move.Retreat {
			move.Retreat = true
			move.Origin, move.Destination = move.Destination, move.Origin
			move.SecondDestination = ""
			origin.incomingMoves = append(origin.incomingMoves, move)
			origin.order.Clear()
			origin.removeUnit()
		}

		if move.hasKnightMove() {
			board[move.SecondDestination].expectedKnightMoves--
		}
	}
}

func (board Board) unitCounts(faction PlayerFaction) (unitCount int, maxUnitCount int) {
	homeRegionsControlled := 0
	var nationsControlled set.ArraySet[string]
	var nationsNotControlled set.ArraySet[string]

	for _, region := range board {
		if region.Unit.HasValue() && region.Unit.Value.Faction == faction {
			unitCount++
		}

		if region.HomeFaction == faction {
			if region.ControllingFaction == faction {
				homeRegionsControlled++
			}
		} else {
			if region.ControllingFaction == faction {
				if !nationsNotControlled.Contains(region.Nation) {
					nationsControlled.Add(region.Nation)
				}
			} else {
				nationsNotControlled.Add(region.Nation)
				nationsControlled.Remove(region.Nation)
			}
		}
	}

	maxUnitCount = homeRegionsControlled + nationsControlled.Size()
	return unitCount, maxUnitCount
}

func (board Board) copy() Board {
	boardCopy := make(Board, len(board))
	for regionName, region := range board {
		regionCopy := *region
		boardCopy[regionName] = &regionCopy
	}
	return boardCopy
}

// Checks whether the region contains a unit.
func (region *Region) empty() bool {
	return region.Unit.IsEmpty()
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
		region.Unit.Clear()
		region.SiegeCount = 0
	}
}

func (region *Region) replaceUnit(unit Unit) {
	region.Unit.Put(unit)
	region.SiegeCount = 0
}

func (region *Region) resolveRetreat() {
	if region.unresolvedRetreat.HasValue() {
		if region.empty() {
			region.Unit.Put(region.unresolvedRetreat.Value.unit())
		}
		region.unresolvedRetreat.Clear()
	}
}

// Returns a region's neighbor of the given name, and whether it was found.
// If the region has several neighbor relations to the region, returns the one matching the provided
// 'via' string.
func (region *Region) getNeighbor(neighborName RegionName, via DangerZone) (Neighbor, bool) {
	var neighbor Neighbor
	hasNeighbor := false

	for _, candidate := range region.Neighbors {
		if neighborName != candidate.Name {
			continue
		}

		if !hasNeighbor {
			neighbor = candidate
			hasNeighbor = true
		} else if candidate.DangerZone != "" && via == candidate.DangerZone {
			neighbor = candidate
		}
	}

	return neighbor, hasNeighbor
}

// Returns whether the region is adjacent to a region of the given name.
func (region *Region) adjacentTo(neighborName RegionName) bool {
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
