package game

import "sync"

type Game struct {
	Board   Board
	Rounds  []*Round
	Players []PlayerColor
}

type PlayerColor string

const Uncontrolled PlayerColor = ""

type Round struct {
	mut          sync.Mutex
	Season       Season
	FirstOrders  []*Order
	SecondOrders []*Order
}

type Season string

const (
	Winter Season = "winter"
	Spring Season = "spring"
	Summer Season = "summer"
	Fall   Season = "fall"
)

type Board map[string]*BoardArea

type BoardArea struct {
	Name             string
	Control          PlayerColor
	Home             PlayerColor
	Unit             *Unit
	Sea              bool
	Forest           bool
	Castle           bool
	SiegeCount       int
	Neighbors        []Neighbor
	IncomingMoves    []*Order
	IncomingSupports []*Order
	Outgoing         *Order
	Combats          []Combat
}

type Neighbor struct {
	Area       *BoardArea
	River      bool
	Cliffs     bool
	DangerZone string
}

type Unit struct {
	Type  UnitType
	Color PlayerColor
}

type UnitType string

const (
	Footman  UnitType = "footman"
	Horse    UnitType = "horse"
	Ship     UnitType = "ship"
	Catapult UnitType = "catapult"
)

type Order struct {
	Type   OrderType
	Player PlayerColor
	From   *BoardArea
	To     *BoardArea
	Via    string
	Build  UnitType
	Status OrderStatus
}

type OrderType string

const (
	Move      OrderType = "move"
	Support   OrderType = "support"
	Transport OrderType = "transport"
	Besiege   OrderType = "besiege"
	Build     OrderType = "build"
)

type OrderStatus string

const (
	Pending OrderStatus = ""
	Success OrderStatus = "success"
	Tie     OrderStatus = "tie"
	Fail    OrderStatus = "fail"
	Error   OrderStatus = "error"
)

type Combat []Result

type Result struct {
	Total  int
	Parts  []Modifier
	Player PlayerColor
}

type Modifier struct {
	Type        ModifierType
	Value       int
	SupportFrom PlayerColor
}

type ModifierType string

const (
	DiceMod     ModifierType = "dice"
	UnitMod     ModifierType = "unit"
	ForestMod   ModifierType = "forest"
	CastleMod   ModifierType = "castle"
	WaterMod    ModifierType = "water"
	SurpriseMod ModifierType = "surprise"
	SupportMod  ModifierType = "support"
)
