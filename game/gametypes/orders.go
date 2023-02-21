package gametypes

// An order submitted by a player for one of their units in a given round.
type Order struct {
	// The type of order submitted. Restricted by unit type and region.
	Type OrderType `json:"type"`

	// The player submitting the order.
	Player string `json:"player"`

	// The unit the order affects.
	// Excluded from JSON messages, as clients can deduce this from the From field.
	// Server includes this field on the order to keep track of units between battles.
	Unit Unit `json:"-"`

	// Name of the region where the order is placed.
	From string `json:"from"`

	// For move and support orders: name of destination region.
	To string `json:"to"`

	// For move orders: name of DangerZone the order tries to pass through, if any.
	Via string `json:"via"`

	// For build orders: type of unit to build.
	Build UnitType `json:"build"`
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
