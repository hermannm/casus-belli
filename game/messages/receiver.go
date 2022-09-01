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
	Orders     chan submitOrdersMsg
	Support    map[string]chan giveSupportMsg
	WinterVote chan winterVoteMsg
	Sword      chan swordMsg
	Raven      chan ravenMsg
}

// Takes a message ID and an unserialized JSON message.
// Unmarshals the message according to its type, and send it to the appropraite receiver channel.
func (receiver Receiver) ReceiveMessage(msgID string, rawMsg json.RawMessage) {
	var err error // Error declared here in order to handle it after the switch.

	switch msgID {
	case submitOrdersMsgID:
		var msg submitOrdersMsg
		if err = json.Unmarshal(rawMsg, &msg); err == nil {
			receiver.Orders <- msg
			return
		}
	case giveSupportMsgID:
		var msg giveSupportMsg
		if err = json.Unmarshal(rawMsg, &msg); err == nil {
			supportChan, ok := receiver.Support[msg.From]
			if ok {
				supportChan <- msg
			} else {
				err = fmt.Errorf("support receiver uninitialized for area %s", msg.From)
			}
		}
	case winterVoteMsgID:
		var msg winterVoteMsg
		if err = json.Unmarshal(rawMsg, &msg); err == nil {
			receiver.WinterVote <- msg
			return
		}
	case swordMsgID:
		var msg swordMsg
		if err = json.Unmarshal(rawMsg, &msg); err == nil {
			receiver.Sword <- msg
			return
		}
	case ravenMsgID:
		var msg ravenMsg
		if err = json.Unmarshal(rawMsg, &msg); err == nil {
			receiver.Raven <- msg
			return
		}
	}

	if err != nil {
		log.Println(fmt.Errorf("failed to parse message of type %s: %w", msgID, err))
	}
}

func (messenger Messenger) ReceiveOrders(from string) ([]board.Order, error) {
	receiver, ok := messenger.receivers[from]
	if !ok {
		return nil, fmt.Errorf("failed to get order message from player %s: receiver not found", from)
	}

	orders := <-receiver.Orders
	return orders.Orders, nil
}

func (messenger Messenger) ReceiveSupport(from string, supportingArea string) (supportTo string, err error) {
	receiver, ok := messenger.receivers[from]
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
