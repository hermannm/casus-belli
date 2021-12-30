package serializer

const (
	OrdersMessageType     = "orders"
	SupportMessageType    = "support"
	QuitMessageType       = "quit"
	WinterVoteMessageType = "winterVote"
)

type IncomingMessages struct {
	Orders     chan OrdersMessage
	Support    chan SupportMessage
	Quit       chan QuitMessage
	WinterVote chan WinterVoteMessage
}

type BaseMessage struct {
	Type string `json:"type"`
}

func (message BaseMessage) GetType() string {
	return message.Type
}

type Message interface {
	GetType() string
}

// Message passed from the client when submitting orders.
type OrdersMessage struct {
	BaseMessage
	Orders []OrderMessage `json:"orders"`
}

type OrderMessage struct {
	Type  string `json:"type"`
	From  string `json:"from"`
	To    string `json:"to"`
	Via   string `json:"via"`
	Build string `json:"build"`
}

// Message passed from the client when declaring who to support with their support order.
type SupportMessage struct {
	BaseMessage
	From   string `json:"from"`   // Name of the area in which the support order is placed.
	To     string `json:"to"`     // Name of the area to which the support order is going.
	Player string `json:"player"` // ID of the player in the destination area to support.
}

type QuitMessage struct {
	BaseMessage
}

// Message passed from the client during winter voting in games that have that enabled.
type WinterVoteMessage struct {
	BaseMessage
	Vote string `json:"vote"` // ID of the player that the submitting player votes for.
}
