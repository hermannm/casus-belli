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

type OrderType string

const (
	Move      OrderType = "move"
	Support   OrderType = "support"
	Transport OrderType = "transport"
	Besiege   OrderType = "besiege"
)
