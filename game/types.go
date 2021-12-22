package game

import "sync"

type Game struct {
	Board   Board
	Rounds  []*Round
	Players []Player
}

type Player string

type Round struct {
	mut          sync.Mutex
	Season       Season
	FirstOrders  []*Order
	SecondOrders []*Order
}

type Season string

type Board map[string]*BoardArea

type BoardArea struct {
	Name             string
	Forest           bool
	Castle           bool
	Sea              bool
	Home             Player
	Control          Player
	Unit             Unit
	SiegeCount       int
	Combats          []Combat
	Neighbors        []Neighbor
	Order            *Order
	IncomingMoves    []*Order
	IncomingSupports []*Order
}

type Neighbor struct {
	Area       *BoardArea
	River      bool
	Cliffs     bool   // Whether coast between neighboring land areas have cliffs (and thus is impassable to ships).
	DangerZone string // If not "": the name of the danger zone that the neighboring area lies across (requires check to pass).
}

type Unit struct {
	Type   UnitType
	Player Player
}

type UnitType string

type Order struct {
	Type   OrderType
	Player Player
	From   *BoardArea
	To     *BoardArea
	Via    string
	Build  UnitType
	Status OrderStatus
}

type OrderType string

type OrderStatus string

type Combat []Result

type Result struct {
	Total  int
	Parts  []Modifier
	Player Player
}

type Modifier struct {
	Type        ModifierType
	Value       int
	SupportFrom Player
}

type ModifierType string

const Uncontrolled Player = ""

const (
	Winter Season = "winter"
	Spring Season = "spring"
	Summer Season = "summer"
	Fall   Season = "fall"
)

const (
	NoUnit   UnitType = ""
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
