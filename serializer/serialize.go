package serializer

import (
	"encoding/json"
	"errors"
)

func Receive(receiver chan []byte, incoming IncomingMessages, errs chan error) {
	for {
		message := <-receiver
		go Deserialize(message, incoming, errs)
	}
}

func Deserialize(rawMessage []byte, incoming IncomingMessages, errs chan error) {
	var baseMessage BaseMessage

	err := json.Unmarshal(rawMessage, &baseMessage)
	if err != nil {
		errs <- err
		return
	}
	if baseMessage.Type == "" {
		errs <- errors.New("error in deserializing message")
		return
	}

	switch baseMessage.Type {

	case OrdersMessageType:
		var ordersMessage OrdersMessage
		err := json.Unmarshal(rawMessage, &ordersMessage)
		if err != nil {
			errs <- err
			return
		}

		incoming.Orders <- ordersMessage

	case SupportMessageType:
		var supportMessage SupportMessage
		err := json.Unmarshal(rawMessage, &supportMessage)
		if err != nil {
			errs <- err
			return
		}

		incoming.Support <- supportMessage

	case QuitMessageType:
		var quitMessage QuitMessage
		err := json.Unmarshal(rawMessage, &quitMessage)
		if err != nil {
			errs <- err
			return
		}

		incoming.Quit <- quitMessage

	case WinterVoteMessageType:
		var winterVoteMessage WinterVoteMessage
		err := json.Unmarshal(rawMessage, &winterVoteMessage)
		if err != nil {
			errs <- err
			return
		}

		incoming.WinterVote <- winterVoteMessage

	}
}
