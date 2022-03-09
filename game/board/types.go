package board

// Unique tag for a player in the game.
type Player string

// A set of player-submitted orders for a round of the game.
type Round struct {
	Season       Season  // Affects the type of orders that can be played in the round.
	FirstOrders  []Order // The main set of orders for the round.
	SecondOrders []Order // Set of orders that are known to be executed after the first orders (e.g. horse moves).
}

// Current season of a round (affects the type of orders that can be played).
// See Season constants for possible values.
type Season string

// A map of area names to areas.
type Board map[string]Area

// An area on the board map.
type Area struct {
	Name      string     // Name of the area on the board.
	Neighbors []Neighbor // Adjacent areas.

	Sea bool // Whether the area is a sea area that can only have ship units.

	Forest bool   // For land areas: affects the difficulty of conquering the area.
	Castle bool   // For land areas: affects the difficulty of conquering the area, and the points gained from it.
	Nation string // For land areas: the collection of areas that the area belongs to (affects units gained from conquering).
	Home   Player // For land areas that are a starting area for a player.

	Unit       Unit   // The unit that currently occupies the area.
	Control    Player // The player that currently controls the area.
	SiegeCount int    // For land areas with castles: the number of times an occupying unit has besieged the castle.

	Order            Order   // Order for the occupying unit in the area. Resets every round.
	IncomingMoves    []Order // Incoming move orders to the area. Resets every round.
	IncomingSupports []Order // Incoming support orders to the area. Resets every round.
}

// The relationship between two adjacent areas.
type Neighbor struct {
	Name        string // Name of the adjacent area.
	AcrossWater bool   // Whether a river separates the two areas.
	Cliffs      bool   // Whether coast between neighboring land areas have cliffs (impassable to ships).
	DangerZone  string // If not "": the name of the danger zone that the neighboring area lies across (requires check to pass).
}

// A player unit on the board.
type Unit struct {
	Type   UnitType // Affects how the unit moves and its battle capabilities.
	Player Player   // The player owning the unit.
}

// Type of player unit on the board (affects how it moves and its battle capabilities).
// See UnitType constants for possible values.
type UnitType string

// An order submitted by a player for one of their units in a given round.
type Order struct {
	Type   OrderType // The type of order submitted. Restricted by unit type and area.
	Player Player    // The player submitting the order.
	Unit   Unit      // The unit the order affects.
	From   string    // Name of the area where the order is placed.

	To string // For move and support orders: name of destination area.

	Via string // For move orders: name of DangerZone the order tries to pass through, if any.

	Build UnitType // For build orders: type of unit to build.
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
	Results []Result

	DangerZone string // In case of danger zone crossing: name of the danger zone.
}

// Dice and modifier result for a battle.
type Result struct {
	Total        int        // The sum of the dice roll and modifiers.
	Parts        []Modifier // The modifiers comprising the result, including the dice roll.
	Move         Order      // If result of a move order to the battle: the move order in question.
	DefenderArea string     // If result of a defending unit in an area: the name of the area.
}

// A typed number that adds to a player's result in a battle.
type Modifier struct {
	Type        ModifierType // The source of the modifier.
	Value       int          // The positive or negative number that modifies the result total.
	SupportFrom Player       // If modifier was from a support: the supporting player.
}

// The source of a modifier.
type ModifierType string

// Number to beat when attempting to conquer a neutral area.
const ConquerRequirement int = 4

// Number to beat when attempting to cross a danger zone.
const DangerZoneRequirement int = 3

// Valid values for a round's season.
const (
	Winter Season = "winter"
	Spring Season = "spring"
	Summer Season = "summer"
	Fall   Season = "fall"
)

// Valid values for a player unit's type.
const (
	Footman  UnitType = "footman"
	Horse    UnitType = "horse"
	Ship     UnitType = "ship"
	Catapult UnitType = "catapult"
)

// Valid values for a player-submitted order's type.
const (
	Move      OrderType = "move"
	Support   OrderType = "support"
	Transport OrderType = "transport"
	Besiege   OrderType = "besiege"
	Build     OrderType = "build"
)

// Valid values for a result modifier's type.
const (
	DiceMod     ModifierType = "dice"
	UnitMod     ModifierType = "unit"
	ForestMod   ModifierType = "forest"
	CastleMod   ModifierType = "castle"
	WaterMod    ModifierType = "water"
	SurpriseMod ModifierType = "surprise"
	SupportMod  ModifierType = "support"
)
