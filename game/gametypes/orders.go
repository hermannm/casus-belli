package gametypes

import "encoding/json"

// An order submitted by a player for one of their units in a given round.
type Order struct {
	// The type of order submitted. Restricted by unit type and region.
	Type OrderType `json:"type"`

	// The player submitting the order.
	Player string `json:"player"`

	// Name of the region where the order is placed.
	Origin string `json:"origin"`

	// For move and support orders: name of destination region.
	Destination string `json:"destination,omitempty"`

	// For move orders with horse units: optional name of second destination region to move to if
	// the first destination was reached.
	SecondDestination string `json:"secondDestination,omitempty"`

	// For move orders: name of DangerZone the order tries to pass through, if any.
	Via string `json:"via,omitempty"`

	// For build orders: type of unit to build.
	Build UnitType `json:"build,omitempty"`

	// The unit the order affects.
	// Excluded from JSON messages, as clients can deduce this from the From field.
	// Server includes this field on the order to keep track of units between battles.
	Unit Unit `json:"-"`
}

// Type of submitted order (restricted by unit type and region).
// See OrderType constants for possible values.
type OrderType string

// Valid values for a player-submitted order's type.
const (
	// An order for a unit to move from one region to another.
	// Includes internal moves in winter.
	OrderMove OrderType = "move"

	// An order for a unit to support battle in an adjacent region.
	OrderSupport OrderType = "support"

	// For ship unit at sea: an order to transport a land unit across the sea.
	OrderTransport OrderType = "transport"

	// For land unit in unconquered castle region: an order to besiege the castle.
	OrderBesiege OrderType = "besiege"

	// For player-controlled region in winter: an order for the type of unit to build in the region.
	OrderBuild OrderType = "build"
)

// Checks whether the order is initialized.
func (order Order) IsNone() bool {
	return order.Type == ""
}

// Checks if the order is a move of a horse unit with a second destination.
// If it is, returns the order with the original destination set as the origin, and the destination
// set as the original second destination.
// Otherwise, returns hasSecondHorseMove=false.
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
