package messages

import (
	"fmt"
)

type Messenger struct {
	sender    Sender
	receivers map[string]Receiver
}

func NewMessenger(sender Sender) Messenger {
	receivers := make(map[string]Receiver)
	return Messenger{
		sender:    sender,
		receivers: receivers,
	}
}

func (messenger Messenger) AddReceiver(playerID string, areaNames []string) (Receiver, error) {
	_, exists := messenger.receivers[playerID]
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

	messenger.receivers[playerID] = receiver
	return receiver, nil
}

func (messenger Messenger) RemoveReceiver(playerID string) {
	delete(messenger.receivers, playerID)
}

func (messenger Messenger) ReceiverIDs() []string {
	ids := make([]string, 0)
	for id := range messenger.receivers {
		ids = append(ids, id)
	}
	return ids
}
