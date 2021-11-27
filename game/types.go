package game

type Player struct {
	ConnectionID string
	Color        PlayerColor
	Units        []*Unit
}

type Unit struct {
	Type  UnitType
	Color PlayerColor
}

type Round struct {
	Board        Board
	FirstOrders  []*Order
	SecondOrders []*Order
}

type Board map[string]*BoardArea

type BoardArea struct {
	Name             string
	Control          PlayerColor
	Unit             *Unit
	Forest           bool
	Castle           bool
	Sea              bool
	Neighbors        map[string]*Neighbor
	IncomingMoves    map[string]*Order
	IncomingSupports map[string]*Order
	Outgoing         *Order
	Combats          []Combat
}

type Neighbor struct {
	Area        *BoardArea
	AcrossWater bool
	DangerZone  string
}

type Order struct {
	Type      OrderType
	Player    *Player
	From      *BoardArea
	To        *BoardArea
	UnitBuild UnitType
	Status    OrderStatus
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
