package messages

import (
	"fmt"
	"log"

	"hermannm.dev/bfh-server/game/board"
)

type Sender interface {
	SendMessage(to string, msg map[string]any) error
	SendMessageToAll(msg map[string]any) error
}

func (messenger Messenger) SendError(to string, errMsg string) {
	err := messenger.sender.SendMessage(to, message{errorMsgID: errorMsg{Error: errMsg}})
	if err != nil {
		log.Println(fmt.Errorf("failed to send error message to player %s: %w", to, err))
	}
}

func (messenger Messenger) SendOrderRequest(to string) error {
	err := messenger.sender.SendMessage(to, message{orderRequestMsgID: orderRequestMsg{}})
	if err != nil {
		return fmt.Errorf("failed to send order request message to player %s: %w", to, err)
	}
	return nil
}

func (messenger Messenger) SendOrdersReceived(playerOrders map[string][]board.Order) error {
	err := messenger.sender.SendMessageToAll(message{ordersReceivedMsgID: ordersReceivedMsg{Orders: playerOrders}})
	if err != nil {
		return fmt.Errorf("failed to send orders received message: %w", err)
	}
	return nil
}

func (messenger Messenger) SendOrdersConfirmation(player string) error {
	err := messenger.sender.SendMessageToAll(message{ordersConfirmationMsgID: ordersConfirmationMsg{Player: player}})
	if err != nil {
		return fmt.Errorf("failed to send orders confirmation message: %w", err)
	}
	return nil
}

func (messenger Messenger) SendSupportRequest(to string, supportingArea string, battlers []string) error {
	err := messenger.sender.SendMessage(to, message{
		supportRequestMsgID: supportRequestMsg{SupportingArea: supportingArea, Battlers: battlers},
	})
	if err != nil {
		return fmt.Errorf("failed to send support request message to player %s: %w", to, err)
	}
	return nil
}

func (messenger Messenger) SendBattleResults(battles []board.Battle) error {
	err := messenger.sender.SendMessageToAll(message{battleResultsMsgID: battleResultsMsg{Battles: battles}})
	if err != nil {
		return fmt.Errorf("failed to send battle results message: %w", err)
	}
	return nil
}

func (messenger Messenger) SendWinner(winner string) error {
	err := messenger.sender.SendMessageToAll(message{winnerMsgID: winnerMsg{Winner: winner}})
	if err != nil {
		return fmt.Errorf("failed to send winner message: %w", err)
	}
	return nil
}
