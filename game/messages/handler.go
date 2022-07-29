package messages

import (
	"fmt"
)

type Handler struct {
	sender    Sender
	receivers map[string]Receiver
}

func NewHandler(sender Sender) Handler {
	receivers := make(map[string]Receiver)
	return Handler{
		sender:    sender,
		receivers: receivers,
	}
}

func (handler Handler) AddReceiver(playerID string, areaNames []string) (Receiver, error) {
	_, exists := handler.receivers[playerID]
	if exists {
		return Receiver{}, fmt.Errorf("message receiver for player id %s already exists", playerID)
	}

	supportChans := make(map[string]chan giveSupportMsg)
	for _, areaName := range areaNames {
		supportChans[areaName] = make(chan giveSupportMsg)
	}

	receiver := Receiver{
		Orders:     make(chan submitOrdersMsg),
		Support:    supportChans,
		WinterVote: make(chan winterVoteMsg),
		Sword:      make(chan swordMsg),
		Raven:      make(chan ravenMsg),
	}

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
