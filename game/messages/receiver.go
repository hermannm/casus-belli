package messages

import (
	"encoding/json"
	"errors"
	"fmt"
)

// The Receiver handles messages coming from client, and parses them
// to the appropriate message type channel for use by the game instance.
type Receiver struct {
	Orders     chan SubmitOrders
	Support    map[string]chan GiveSupport
	Quit       chan Quit
	Kick       chan Kick
	WinterVote chan WinterVote
	Sword      chan Sword
	Raven      chan Raven
	Errors     chan error
}

// Initializes a new receiver with empty channels.
func NewReceiver(areaNames []string) Receiver {
	supportChans := make(map[string]chan GiveSupport)
	for _, areaName := range areaNames {
		supportChans[areaName] = make(chan GiveSupport)
	}

	return Receiver{
		Orders:     make(chan SubmitOrders),
		Support:    supportChans,
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
func (receiver Receiver) ReceiveMessage(msgType string, rawMsg []byte) {
	switch msgType {
	case MsgSubmitOrders:
		var ordersMessage SubmitOrders
		err := json.Unmarshal(rawMsg, &ordersMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Orders <- ordersMessage
	case MsgGiveSupport:
		var supportMessage GiveSupport
		err := json.Unmarshal(rawMsg, &supportMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Support[supportMessage.From] <- supportMessage
	case MsgQuit:
		var quitMessage Quit
		err := json.Unmarshal(rawMsg, &quitMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Quit <- quitMessage
	case MsgKick:
		var kickMessage Kick
		err := json.Unmarshal(rawMsg, &kickMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Kick <- kickMessage
	case MsgWinterVote:
		var winterVoteMessage WinterVote
		err := json.Unmarshal(rawMsg, &winterVoteMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.WinterVote <- winterVoteMessage
	case MsgSword:
		var swordMessage Sword
		err := json.Unmarshal(rawMsg, &swordMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Sword <- swordMessage
	case MsgRaven:
		var ravenMessage Raven
		err := json.Unmarshal(rawMsg, &ravenMessage)
		if err != nil {
			receiver.Errors <- err
			return
		}

		receiver.Raven <- ravenMessage
	default:
		receiver.Errors <- errors.New("unrecognized message type: " + msgType)
		return
	}
}

func (handler Handler) AddReceiver(playerID string, areaNames []string) (Receiver, error) {
	_, exists := handler.receivers[playerID]
	if exists {
		return Receiver{}, fmt.Errorf("message receiver for player id %s already exists", playerID)
	}

	receiver := NewReceiver(areaNames)
	handler.receivers[playerID] = receiver
	return receiver, nil
}

func (handler Handler) RemoveReceiver(playerID string) {
	delete(handler.receivers, playerID)
}

func (handler Handler) ReceiverIDs() []string {
	ids := make([]string, 0)
	for id := range handler.receivers {
		ids = append(ids, id)
	}
	return ids
}
