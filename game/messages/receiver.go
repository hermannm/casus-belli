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
// Unmarshals the message according to its type, and sends it to the appropraite receiver channel.
func (receiver Receiver) ReceiveMessage(msgID string, rawMsg json.RawMessage) {
	var err error // Error declared here in order to handle it after the switch.

	switch msgID {
	case submitOrdersMsgID:
		var msg submitOrdersMsg
		err = json.Unmarshal(rawMsg, &msg)
		if err == nil {
			receiver.Orders <- msg
		}
	case giveSupportMsgID:
		var msg giveSupportMsg
		err = json.Unmarshal(rawMsg, &msg)
		if err == nil {
			supportChan, ok := receiver.Support[msg.SupportingRegion]
			if ok {
				supportChan <- msg
			} else {
				err = fmt.Errorf("support receiver uninitialized for region %s", msg.SupportingRegion)
			}
		}
	case winterVoteMsgID:
		var msg winterVoteMsg
		err = json.Unmarshal(rawMsg, &msg)
		if err == nil {
			receiver.WinterVote <- msg
		}
	case swordMsgID:
		var msg swordMsg
		err = json.Unmarshal(rawMsg, &msg)
		if err == nil {
			receiver.Sword <- msg
		}
	case ravenMsgID:
		var msg ravenMsg
		err = json.Unmarshal(rawMsg, &msg)
		if err == nil {
			receiver.Raven <- msg
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

func (messenger Messenger) ReceiveSupport(from string, supportingRegion string) (supportTo string, err error) {
	receiver, ok := messenger.receivers[from]
	if !ok {
		return "", fmt.Errorf("failed to get support message from player %s: receiver not found", from)
	}

	supportChan, ok := receiver.Support[supportingRegion]
	if !ok {
		return "", fmt.Errorf(
			"failed to get support message from player %s: support receiver uninitialized for region %s",
			from,
			supportingRegion,
		)
	}

	support := <-supportChan

	if support.SupportedPlayer != nil {
		return *support.SupportedPlayer, nil
	} else {
		return "", nil
	}
}
