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
	Yellow       PlayerColor = "yellow"
	Red          PlayerColor = "red"
	Green        PlayerColor = "green"
	White        PlayerColor = "white"
	Black        PlayerColor = "black"
	Uncontrolled PlayerColor = ""
)

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
	Success OrderStatus = "success"
	Tie     OrderStatus = "tie"
	Fail    OrderStatus = "fail"
	Pending OrderStatus = "pending"
	Error   OrderStatus = "error"
)

type Season string

const (
	Winter Season = "winter"
	Spring Season = "spring"
	Summer Season = "summer"
	Fall   Season = "fall"
)

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
