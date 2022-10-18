package board

type Messenger interface {
	SendBattleResults(battles []Battle) error
	SendSupportRequest(to string, supportingArea string, battlers []string) error
	ReceiveSupport(from string, fromArea string) (supportTo string, err error)
}

// A set of player-submitted orders for a round of the game.
type Round struct {
	// Affects the type of orders that can be played in the round.
	Season Season `json:"season"`

	// The main set of orders for the round.
	FirstOrders []Order `json:"firstOrders"`

	// Set of orders that are known to be executed after the first orders (e.g. horse moves).
	SecondOrders []Order `json:"secondOrders"`
}

// Current season of a round (affects the type of orders that can be played).
// See Season constants for possible values.
type Season string

// A pre-configured board used for the game.
type Board struct {
	// Areas on the board that player units can move to.
	Areas map[string]Area `json:"area"`

	// Name of this type of board.
	Name string

	// The number of castles to capture to win a game round on this board.
	WinningCastleCount int `json:"winningCastleCount"`
}

// An area on the board map.
type Area struct {
	// Name of the area on the board.
	Name string `json:"name"`

	// Adjacent areas.
	Neighbors []Neighbor `json:"neighbors"`

	// Whether the area is a sea area that can only have ship units.
	Sea bool `json:"sea"`

	// For land areas: affects the difficulty of conquering the area.
	Forest bool `json:"forest"`

	// For land areas: affects the difficulty of conquering the area, and the points gained from it.
	Castle bool `json:"castle"`

	// For land areas: the collection of areas that the area belongs to (affects units gained from conquering).
	Nation string `json:"nation"`

	// For land areas that are a starting area for a player.
	HomePlayer string `json:"homePlayer,omitempty"`

	// The unit that currently occupies the area.
	Unit Unit `json:"unit"`

	// The player that currently controls the area.
	ControllingPlayer string `json:"controllingPlayer,omitempty"`

	// For land areas with castles: the number of times an occupying unit has besieged the castle.
	SiegeCount int `json:"siegeCount"`

	// Order for the occupying unit in the area. Resets every round.
	Order Order `json:"-"` // Excluded from JSON responses.

	// Incoming move orders to the area. Resets every round.
	IncomingMoves []Order `json:"-"` // Excluded from JSON responses.

	// Incoming support orders to the area. Resets every round.
	IncomingSupports []Order `json:"-"` // Excluded from JSON responses.
}

// The relationship between two adjacent areas.
type Neighbor struct {
	// Name of the adjacent area.
	Name string `json:"name"`

	// Whether a river separates the two areas.
	AcrossWater bool `json:"acrossWater"`

	// Whether coast between neighboring land areas have cliffs (impassable to ships).
	Cliffs bool `json:"cliffs"`

	// If not "": the name of the danger zone that the neighboring area lies across (requires check to pass).
	DangerZone string `json:"dangerZone"`
}

// A player unit on the board.
type Unit struct {
	// Affects how the unit moves and its battle capabilities.
	Type UnitType `json:"unit"`

	// The player owning the unit.
	Player string `json:"player"`
}

// Type of player unit on the board (affects how it moves and its battle capabilities).
// See UnitType constants for possible values.
type UnitType string

// An order submitted by a player for one of their units in a given round.
type Order struct {
	// The type of order submitted. Restricted by unit type and area.
	Type OrderType `json:"type"`

	// The player submitting the order.
	Player string `json:"player"`

	// The unit the order affects.
	// Excluded from JSON messages, as clients can deduce this from the From field.
	// Server includes this field on the order to keep track of units between battles.
	Unit Unit `json:"-"`

	// Name of the area where the order is placed.
	From string `json:"from"`

	// For move and support orders: name of destination area.
	To string `json:"to"`

	// For move orders: name of DangerZone the order tries to pass through, if any.
	Via string `json:"via"`

	// For build orders: type of unit to build.
	Build UnitType `json:"build"`
}

// Type of submitted order (restricted by unit type and area).
// See OrderType constants for possible values.
type OrderType string

// Results of a battle from conflicting move orders, an attempt to conquer a neutral area,
// or an attempt to cross a danger zone.
type Battle struct {
	// The dice and modifier results of the battle.
	// If length is one, the battle was a neutral conquer attempt.
	// If length is more than one, the battle was between players.
	Results []Result `json:"results"`

	// In case of danger zone crossing: name of the danger zone.
	DangerZone string `json:"dangerZone"`
}

// Dice and modifier result for a battle.
type Result struct {
	// The sum of the dice roll and modifiers.
	Total int `json:"total"`

	// The modifiers comprising the result, including the dice roll.
	Parts []Modifier `json:"parts"`

	// If result of a move order to the battle: the move order in question.
	Move Order `json:"move"`

	// If result of a defending unit in an area: the name of the area.
	DefenderArea string `json:"defenderArea"`
}

// A typed number that adds to a player's result in a battle.
type Modifier struct {
	// The source of the modifier.
	Type ModifierType `json:"type"`

	// The positive or negative number that modifies the result total.
	Value int `json:"value"`

	// If modifier was from a support: the supporting player.
	SupportingPlayer string `json:"supportingPlayer"`
}

// The source of a modifier.
type ModifierType string

// Numbers to beat in different types of battles.
const (
	// Number to beat when attempting to conquer a neutral area.
	RequirementConquer int = 4

	// Number to beat when attempting to cross a danger zone.
	RequirementDangerZone int = 3
)

// Rounds where only build and internal move orders are allowed.
const SeasonWinter Season = "winter"

// Rounds where only move, support, transport and besiege orders are allowed.
const (
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonFall   Season = "fall"
)

// Valid values for a player unit's type.
const (
	// A land unit that gets a +1 modifier in battle.
	UnitFootman UnitType = "footman"

	// A land unit that moves 2 areas at a time.
	UnitHorse UnitType = "horse"

	// A unit that can move into sea areas and coastal areas.
	UnitShip UnitType = "ship"

	// A land unit that instantly conquers neutral castles, and gets a +1 modifier in attacks on castles.
	UnitCatapult UnitType = "catapult"
)

// Valid values for a player-submitted order's type.
const (
	// An order for a unit to move from one area to another.
	// Includes internal moves in winter.
	OrderMove OrderType = "move"

	// An order for a unit to support battle in an adjacent area.
	OrderSupport OrderType = "support"

	// For ship unit at sea: an order to transport a land unit across the sea.
	OrderTransport OrderType = "transport"

	// For land unit in unconquered castle area: an order to besiege the castle.
	OrderBesiege OrderType = "besiege"

	// For player-controlled area in winter: an order for what type of unit to build in the area.
	OrderBuild OrderType = "build"
)

// Valid values for a result modifier's type.
const (
	// Bonus from a random dice roll.
	ModifierDice ModifierType = "dice"

	// Bonus for the type of unit.
	ModifierUnit ModifierType = "unit"

	// Penalty for attacking a neutral or defended forested area.
	ModifierForest ModifierType = "forest"

	// Penalty for attacking a neutral or defended castle area.
	ModifierCastle ModifierType = "castle"

	// Penalty for attacking across a river, from the sea, or across a transport.
	ModifierWater ModifierType = "water"

	// Bonus for attacking across a danger zone and surviving.
	ModifierSurprise ModifierType = "surprise"

	// Bonus from supporting player in a battle.
	ModifierSupport ModifierType = "support"
)
