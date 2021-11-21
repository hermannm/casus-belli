package types

type Player struct {
	ConnectionID string
	Color        PlayerColor
	Units        []Unit
}

type Unit struct {
	Type  UnitType
	Color PlayerColor
}

type BoardArea struct {
	Name              string
	ControllingPlayer PlayerColor
	OccupyingUnit     Unit
	Forest            bool
	Castle            bool
	Sea               bool
	Neighbors         map[string]Neighbor
}

type Neighbor struct {
	Area        BoardArea
	AcrossWater bool
}

type Order struct {
	Type     OrderType
	From     BoardArea
	To       BoardArea
	SecondTo BoardArea
}
