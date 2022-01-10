package messages

import (
	"encoding/json"
	"errors"
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

// Takes a message in byte format, deserializes it, and sends it to the appropriate channel on the receiver.
func (receiver *Receiver) HandleMessage(rawMessage []byte) {
	var baseMessage Base

	err := json.Unmarshal(rawMessage, &baseMessage)
	if err != nil {
		receiver.Errors <- err
		return
	}
	if baseMessage.Type == "" {
		receiver.Errors <- errors.New("error in deserializing message")
		return
	}

	switch baseMessage.Type {

	case SubmitOrdersType:
		var ordersMessage SubmitOrders
		err := json.Unmarshal(rawMessage, &ordersMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Orders <- ordersMessage

	case GiveSupportType:
		var supportMessage GiveSupport
		err := json.Unmarshal(rawMessage, &supportMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Support <- supportMessage

	case QuitType:
		var quitMessage Quit
		err := json.Unmarshal(rawMessage, &quitMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Quit <- quitMessage

	case KickType:
		var kickMessage Kick
		err := json.Unmarshal(rawMessage, &kickMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Kick <- kickMessage

	case WinterVoteType:
		var winterVoteMessage WinterVote
		err := json.Unmarshal(rawMessage, &winterVoteMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.WinterVote <- winterVoteMessage

	case SwordType:
		var swordMessage Sword
		err := json.Unmarshal(rawMessage, &swordMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Sword <- swordMessage

	case RavenType:
		var ravenMessage Raven
		err := json.Unmarshal(rawMessage, &ravenMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Raven <- ravenMessage

	default:
		receiver.Errors <- errors.New("unrecognized message type")
		return
	}
}
