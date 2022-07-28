package messages

import (
	"fmt"
)

type Handler struct {
	sender    Sender
	receivers map[string]Receiver
}

func NewHandler(s Sender) Handler {
	rs := make(map[string]Receiver)
	return Handler{
		sender:    s,
		receivers: rs,
	}
}

func (h Handler) AddReceiver(playerID string, areaNames []string) (Receiver, error) {
	_, exists := h.receivers[playerID]
	if exists {
		return Receiver{}, fmt.Errorf("message receiver for player id %s already exists", playerID)
	}

	supportChans := make(map[string]chan GiveSupport)
	for _, areaName := range areaNames {
		supportChans[areaName] = make(chan GiveSupport)
	}

	r := Receiver{
		Orders:     make(chan SubmitOrders),
		Support:    supportChans,
		WinterVote: make(chan WinterVote),
		Sword:      make(chan Sword),
		Raven:      make(chan Raven),
	}

	h.receivers[playerID] = r
	return r, nil
}

func (h Handler) RemoveReceiver(playerID string) {
	delete(h.receivers, playerID)
}

func (h Handler) ReceiverIDs() []string {
	ids := make([]string, 0)
	for id := range h.receivers {
		ids = append(ids, id)
	}
	return ids
}
