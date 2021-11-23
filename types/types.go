package types

type Player struct {
	ConnectionID string
	Color        PlayerColor
	Units        []*Unit
}

type Unit struct {
	Type  UnitType
	Color PlayerColor
}

type Board map[string]*BoardArea

type BoardArea struct {
	Name      string
	Control   PlayerColor
	Unit      *Unit
	Forest    bool
	Castle    bool
	Sea       bool
	Neighbors map[string]*Neighbor
	Incoming  []*Order
}

type Neighbor struct {
	Area        *BoardArea
	AcrossWater bool
	DangerZone  string
}

type Order struct {
	Type         OrderType
	Player       *Player
	From         *BoardArea
	To           *BoardArea
	Dependencies []*Order
	UnitBuild    UnitType
	Status       OrderStatus
	Result       CombatResult
}

type CombatResult struct {
	Total int
	Parts []Modifier
}

type Modifier struct {
	Type        ModifierType
	Value       int
	SupportFrom PlayerColor
}
