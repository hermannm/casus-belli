package serializer

import (
	"encoding/json"
	"errors"
)

func Deserialize(rawMessages chan []byte, deserialized IncomingMessages, errs chan error) {
	for rawMessage := range rawMessages {
		var baseMessage BaseMessage

		err := json.Unmarshal(rawMessage, &baseMessage)
		if err != nil {
			errs <- err
			continue
		}
		if baseMessage.Type == "" {
			errs <- errors.New("error in deserializing message")
			continue
		}

		switch baseMessage.Type {

		case OrdersMessageType:
			var ordersMessage OrdersMessage
			err := json.Unmarshal(rawMessage, &ordersMessage)
			if err != nil {
				errs <- err
				continue
			}

			deserialized.Orders <- ordersMessage

		case SupportMessageType:
			var supportMessage SupportMessage
			err := json.Unmarshal(rawMessage, &supportMessage)
			if err != nil {
				errs <- err
				continue
			}

			deserialized.Support <- supportMessage

		case QuitMessageType:
			var quitMessage QuitMessage
			err := json.Unmarshal(rawMessage, &quitMessage)
			if err != nil {
				errs <- err
				continue
			}

			deserialized.Quit <- quitMessage

		case WinterVoteMessageType:
			var winterVoteMessage WinterVoteMessage
			err := json.Unmarshal(rawMessage, &winterVoteMessage)
			if err != nil {
				errs <- err
				continue
			}

			deserialized.WinterVote <- winterVoteMessage

		default:
			errs <- errors.New("unrecognized message type")
			continue
		}
	}
}
