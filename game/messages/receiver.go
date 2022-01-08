package messages

import (
	"encoding/json"
	"errors"
)

type Receiver struct {
	Orders     chan SubmitOrders
	Support    chan GiveSupport
	Quit       chan Quit
	Kick       chan Kick
	WinterVote chan WinterVote
	Errors     chan error
}

func NewReceiver() Receiver {
	return Receiver{
		Orders:     make(chan SubmitOrders),
		Support:    make(chan GiveSupport),
		Quit:       make(chan Quit),
		Kick:       make(chan Kick),
		WinterVote: make(chan WinterVote),
		Errors:     make(chan error),
	}
}

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

	case WinterVoteType:
		var winterVoteMessage WinterVote
		err := json.Unmarshal(rawMessage, &winterVoteMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.WinterVote <- winterVoteMessage

	default:
		receiver.Errors <- errors.New("unrecognized message type")
		return
	}
}
