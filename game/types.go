package game

import (
	"hermannm.dev/bfh-server/lobby"
	"hermannm.dev/bfh-server/messages"
)

type Game struct {
	Board    Board
	Rounds   []Round
	Lobby    *lobby.Lobby
	Messages map[Player]*messages.Receiver
	Options  GameOptions
}

type GameOptions struct {
	Thrones bool // Whether the game has the "Raven, Sword and Throne" expansion enabled.
}

type Player string

type Round struct {
	Season       Season
	FirstOrders  []Order
	SecondOrders []Order
}

type Season string

// A map of area names to areas.
type Board map[string]Area

// An area on the board map.
type Area struct {
	Name      string     // Name of the area on the board.
	Neighbors []Neighbor // Adjacent areas.

	Sea bool // Whether the area is a sea area that can only have ship units.

	Forest bool   // For land areas.
	Castle bool   // For land areas.
	Nation string // For land areas.
	Home   Player // For land areas that are a starting area for a player.

	Unit       Unit   // May change round-to-round.
	Control    Player // May change round-to-round.
	SiegeCount int    // For land areas with castles. May change round-to-round.

	Order            Order   // Order for the unit in the area. Resets every round.
	IncomingMoves    []Order // Incoming move orders to area. Resets every round.
	IncomingSupports []Order // Incoming support orders to area. Resets every round.
}

// The relationship between two adjacent areas.
type Neighbor struct {
	Name        string // Name of the adjacent area.
	AcrossWater bool   // Whether a river separates the two areas.
	Cliffs      bool   // Whether coast between neighboring land areas have cliffs (impassable to ships).
	DangerZone  string // If not "": the name of the danger zone that the neighboring area lies across (requires check to pass).
}

type Unit struct {
	Type   UnitType
	Player Player
}

type UnitType string

type Order struct {
	Type   OrderType
	Player Player
	Unit   Unit
	From   string // Name of area where order is placed.

	To string // For move and support orders: Name of destination area.

	Via string // For move orders: Name of DangerZone the order tries to pass through, if any.

	Build UnitType // For build orders: Name of unit type to build.
}

type OrderType string

type OrderStatus string

type Battle struct {
	Results    []Result
	DangerZone string // In case of danger zone crossing: name of the danger zone.
}

type Result struct {
	Total        int
	Parts        []Modifier
	Move         Order
	DefenderArea string
}

type Modifier struct {
	Type        ModifierType
	Value       int
	SupportFrom Player
}

type ModifierType string

const ConquerRequirement int = 4

const DangerZoneRequirement int = 3

const Uncontrolled Player = ""

const (
	Winter Season = "winter"
	Spring Season = "spring"
	Summer Season = "summer"
	Fall   Season = "fall"
)

const (
	Footman  UnitType = "footman"
	Horse    UnitType = "horse"
	Ship     UnitType = "ship"
	Catapult UnitType = "catapult"
)

const (
	Move      OrderType = "move"
	Support   OrderType = "support"
	Transport OrderType = "transport"
	Besiege   OrderType = "besiege"
	Build     OrderType = "build"
)

const (
	Pending OrderStatus = ""
	Success OrderStatus = "success"
	Tie     OrderStatus = "tie"
	Fail    OrderStatus = "fail"
	Error   OrderStatus = "error"
)

const (
	DiceMod     ModifierType = "dice"
	UnitMod     ModifierType = "unit"
	ForestMod   ModifierType = "forest"
	CastleMod   ModifierType = "castle"
	WaterMod    ModifierType = "water"
	SurpriseMod ModifierType = "surprise"
	SupportMod  ModifierType = "support"
)
