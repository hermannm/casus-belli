package messages

import (
	"encoding/json"
	"errors"

	"hermannm.dev/bfh-server/lobby"
)

// The Receiver handles messages coming from client, and parses them
// to the appropriate message type channel for use by the game instance.
type Receiver struct {
	Orders     chan SubmitOrders
	Support    chan GiveSupport
	Quit       chan Quit
	Kick       chan Kick
	WinterVote chan WinterVote
	Sword      chan Sword
	Raven      chan Raven
	Errors     chan error
}

// Initializes a new receiver with empty channels.
func NewReceiver() Receiver {
	return Receiver{
		Orders:     make(chan SubmitOrders),
		Support:    make(chan GiveSupport),
		Quit:       make(chan Quit),
		Kick:       make(chan Kick),
		WinterVote: make(chan WinterVote),
		Sword:      make(chan Sword),
		Raven:      make(chan Raven),
		Errors:     make(chan error),
	}
}

// Takes a partly deserialized base message, checks it type, and further deserializes the given raw
// message to pass it to the appropriate channel on the receiver.
func (receiver Receiver) HandleMessage(baseMessage lobby.Message, rawMessage []byte) {
	switch baseMessage.Type {
	case MessageSubmitOrders:
		var ordersMessage SubmitOrders
		err := json.Unmarshal(rawMessage, &ordersMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Orders <- ordersMessage
	case MessageGiveSupport:
		var supportMessage GiveSupport
		err := json.Unmarshal(rawMessage, &supportMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Support <- supportMessage
	case MessageQuit:
		var quitMessage Quit
		err := json.Unmarshal(rawMessage, &quitMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Quit <- quitMessage
	case MessageKick:
		var kickMessage Kick
		err := json.Unmarshal(rawMessage, &kickMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Kick <- kickMessage
	case MessageWinterVote:
		var winterVoteMessage WinterVote
		err := json.Unmarshal(rawMessage, &winterVoteMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.WinterVote <- winterVoteMessage
	case MessageSword:
		var swordMessage Sword
		err := json.Unmarshal(rawMessage, &swordMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Sword <- swordMessage
	case MessageRaven:
		var ravenMessage Raven
		err := json.Unmarshal(rawMessage, &ravenMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Raven <- ravenMessage
	default:
		receiver.Errors <- errors.New("unrecognized message type: " + baseMessage.Type)
		return
	}
}
