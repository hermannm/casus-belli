package messages

import (
	"encoding/json"
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/board"
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

// Takes a partly deserialized base message, checks it type, and further deserializes the given raw
// message to pass it to the appropriate channel on the receiver.
func (r Receiver) ReceiveMessage(msgType string, rawMsg []byte) {
	switch msgType {
	case MsgSubmitOrders:
		var ordersMessage SubmitOrders
		err := json.Unmarshal(rawMsg, &ordersMessage)
		if err != nil {
			r.Errors <- err
			return
		}

		r.Orders <- ordersMessage
	case MsgGiveSupport:
		var supportMessage GiveSupport
		err := json.Unmarshal(rawMsg, &supportMessage)
		if err != nil {
			r.Errors <- err
			return
		}

		r.Support[supportMessage.From] <- supportMessage
	case MsgQuit:
		var quitMessage Quit
		err := json.Unmarshal(rawMsg, &quitMessage)
		if err != nil {
			r.Errors <- err
			return
		}

		r.Quit <- quitMessage
	case MsgKick:
		var kickMessage Kick
		err := json.Unmarshal(rawMsg, &kickMessage)
		if err != nil {
			r.Errors <- err
			return
		}

		r.Kick <- kickMessage
	case MsgWinterVote:
		var winterVoteMessage WinterVote
		err := json.Unmarshal(rawMsg, &winterVoteMessage)
		if err != nil {
			r.Errors <- err
			return
		}

		r.WinterVote <- winterVoteMessage
	case MsgSword:
		var swordMessage Sword
		err := json.Unmarshal(rawMsg, &swordMessage)
		if err != nil {
			r.Errors <- err
			return
		}

		r.Sword <- swordMessage
	case MsgRaven:
		var ravenMessage Raven
		err := json.Unmarshal(rawMsg, &ravenMessage)
		if err != nil {
			r.Errors <- err
			return
		}

		r.Raven <- ravenMessage
	default:
		r.Errors <- errors.New("unrecognized message type: " + msgType)
		return
	}
}

func (h Handler) ReceiveOrders(from string) ([]board.Order, error) {
	receiver, ok := h.receivers[from]
	if !ok {
		return nil, fmt.Errorf("failed to get order message from player %s: receiver not found", from)
	}

	orders := <-receiver.Orders
	return orders.Orders, nil
}

func (h Handler) ReceiveSupport(from string, supportingArea string) (supportTo string, err error) {
	receiver, ok := h.receivers[from]
	if !ok {
		return "", fmt.Errorf("failed to get support message from player %s: receiver not found", from)
	}

	supportChan, ok := receiver.Support[supportingArea]
	if !ok {
		return "", fmt.Errorf(
			"failed to get support message from player %s: support receiver uninitialized for area %s",
			from,
			supportingArea,
		)
	}

	support := <-supportChan
	return support.Player, nil
}
