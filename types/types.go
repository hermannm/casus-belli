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

type BoardArea struct {
	Name      string
	Control   PlayerColor
	Unit      *Unit
	Forest    bool
	Castle    bool
	Sea       bool
	Neighbors map[string]*Neighbor
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
}

type OrderResult struct {
	Status OrderStatus
	Order  *Order
	Dice   DieResult
}

type DieResult struct {
	Result    int
	Modifiers []Modifier
}

type Modifier struct {
	Type        ModifierType
	Mod         int
	SupportFrom PlayerColor
}
