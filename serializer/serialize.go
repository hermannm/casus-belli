package serializer

import (
	"encoding/json"
	"errors"
)

func Deserialize(rawMessage []byte, incoming IncomingMessages) error {
	var baseMessage BaseMessage

	err := json.Unmarshal(rawMessage, &baseMessage)
	if err != nil {
		return err
	}
	if baseMessage.Type == "" {
		return errors.New("error in deserializing message")
	}

	switch baseMessage.Type {

	case OrdersMessageType:
		var ordersMessage OrdersMessage
		err := json.Unmarshal(rawMessage, &ordersMessage)
		if err != nil {
			return err
		}

		incoming.Orders <- ordersMessage

	case SupportMessageType:
		var supportMessage SupportMessage
		err := json.Unmarshal(rawMessage, &supportMessage)
		if err != nil {
			return err
		}

		incoming.Support <- supportMessage

	case QuitMessageType:
		var quitMessage QuitMessage
		err := json.Unmarshal(rawMessage, &quitMessage)
		if err != nil {
			return err
		}

		incoming.Quit <- quitMessage

	case WinterVoteMessageType:
		var winterVoteMessage WinterVoteMessage
		err := json.Unmarshal(rawMessage, &winterVoteMessage)
		if err != nil {
			return err
		}

		incoming.WinterVote <- winterVoteMessage

	}

	return nil
}
