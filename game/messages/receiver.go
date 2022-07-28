package messages

import (
	"encoding/json"
	"fmt"
	"log"

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
}

// Takes a partly deserialized base message, checks it type, and further deserializes the given raw
// message to pass it to the appropriate channel on the receiver.
func (r Receiver) ReceiveMessage(msgType string, rawMsg []byte) {
	var err error // Error declared here in order to handle it after the switch.

	switch msgType {
	case MsgSubmitOrders:
		var msg SubmitOrders
		err = json.Unmarshal(rawMsg, &msg)
		if err == nil {
			r.Orders <- msg
			return
		}
	case MsgGiveSupport:
		var msg GiveSupport
		err = json.Unmarshal(rawMsg, &msg)
		if err == nil {
			supportChan, ok := r.Support[msg.From]
			if ok {
				supportChan <- msg
			} else {
				err = fmt.Errorf("support receiver uninitialized for area %s", msg.From)
			}
		}
	case MsgQuit:
		var msg Quit
		err = json.Unmarshal(rawMsg, &msg)
		if err == nil {
			r.Quit <- msg
			return
		}
	case MsgKick:
		var kickMessage Kick
		err = json.Unmarshal(rawMsg, &kickMessage)
		if err == nil {
			r.Kick <- kickMessage
			return
		}
	case MsgWinterVote:
		var winterVoteMessage WinterVote
		err = json.Unmarshal(rawMsg, &winterVoteMessage)
		if err == nil {
			r.WinterVote <- winterVoteMessage
			return
		}
	case MsgSword:
		var swordMessage Sword
		err = json.Unmarshal(rawMsg, &swordMessage)
		if err == nil {
			r.Sword <- swordMessage
			return
		}
	case MsgRaven:
		var ravenMessage Raven
		err = json.Unmarshal(rawMsg, &ravenMessage)
		if err == nil {
			r.Raven <- ravenMessage
			return
		}
	}

	if err != nil {
		log.Println(fmt.Errorf("failed to parse message of type %s: %w", msgType, err))
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
