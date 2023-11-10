package gametypes

import "encoding/json"

// An order submitted by a player for one of their units in a given round.
type Order struct {
	Type    OrderType
	Faction PlayerFaction

	// Name of the region where the order is placed.
	Origin string

	// For move and support orders: name of destination region.
	Destination string

	// For move orders with horse units: optional name of second destination region to move to if
	// the first destination was reached.
	SecondDestination string

	// For move orders: name of DangerZone the order tries to pass through, if any.
	ViaDangerZone string

	// For build orders: type of unit to build.
	Build UnitType

	// The unit the order affects.
	// Excluded from JSON messages, as clients can deduce this from the From field.
	// Server includes this field on the order to keep track of units between battles.
	Unit Unit `json:"-"`
}

type OrderType string

const (
	// An order for a unit to move from one region to another.
	// Includes internal moves in winter.
	OrderMove OrderType = "move"

	// An order for a unit to support battles in adjacent regions.
	OrderSupport OrderType = "support"

	// For ship unit at sea: an order to transport a land unit across the sea.
	OrderTransport OrderType = "transport"

	// For land unit in unconquered castle region: an order to besiege the castle.
	OrderBesiege OrderType = "besiege"

	// For player-controlled region in winter: an order for the type of unit to build in the region.
	OrderBuild OrderType = "build"
)

func (order Order) IsNone() bool {
	return order.Type == ""
}

// Checks if the order is a move of a horse unit with a second destination.
// If it is, returns the order with the original destination set as the origin, and the destination
// set as the original second destination.
func (order Order) TryGetSecondHorseMove() (secondHorseMove Order, hasSecondHorseMove bool) {
	if order.Type != OrderMove || order.SecondDestination == "" || order.Unit.Type != UnitHorse {
		return Order{}, false
	}

	order.Origin = order.Destination
	order.Destination = order.SecondDestination
	order.SecondDestination = ""

	return order, true
}

// Custom json.Marshaler implementation, to serialize uninitialized orders to null.
func (order Order) MarshalJSON() ([]byte, error) {
	if order.IsNone() {
		return []byte("null"), nil
	}

	// Alias to avoid infinite loop of MarshalJSON.
	type orderAlias Order

	return json.Marshal(orderAlias(order))
}
