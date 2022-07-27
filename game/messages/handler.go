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

	receiver := NewReceiver(areaNames)
	handler.receivers[playerID] = receiver
	return receiver, nil
}
