package game

import "sync"

type Game struct {
	Board   Board
	Rounds  []*Round
	Players []*Player
}

type Player struct {
	Color PlayerColor
}

type Round struct {
	mut          sync.Mutex
	Season       Season
	FirstOrders  []*Order
	SecondOrders []*Order
}

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
	IncomingMoves    map[string]*Order
	IncomingSupports map[string]*Order
	Outgoing         *Order
	Combats          []Combat
}

type Unit struct {
	Type  UnitType
	Color PlayerColor
}

type Neighbor struct {
	Area       *BoardArea
	River      bool
	Cliffs     bool
	DangerZone string
}

type Order struct {
	Type   OrderType
	Player *Player
	From   *BoardArea
	To     *BoardArea
	Via    string
	Build  UnitType
	Status OrderStatus
}

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
