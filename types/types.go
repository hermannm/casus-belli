package types

type UnitType string

const (
	Footman  UnitType = "footman"
	Horse    UnitType = "horse"
	Ship     UnitType = "ship"
	Catapult UnitType = "catapult"
)

type PlayerColor string

const (
	Yellow PlayerColor = "yellow"
	Red    PlayerColor = "red"
	Green  PlayerColor = "green"
	White  PlayerColor = "white"
	Black  PlayerColor = "black"
)

type Unit struct {
	Type  UnitType
	Color PlayerColor
}

type Neighbor struct {
	Area        BoardArea
	AcrossWater bool
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

type Order struct {
	To   BoardArea
	From BoardArea
}
