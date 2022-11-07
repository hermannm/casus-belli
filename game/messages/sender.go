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
	msg := message{errorMsgID: errorMsg{Error: errMsg}}

	err := messenger.sender.SendMessage(to, msg)
	if err != nil {
		log.Println(fmt.Errorf("failed to send error message to player %s: %w", to, err))
	}
}

func (messenger Messenger) SendOrderRequest(to string) error {
	msg := message{orderRequestMsgID: orderRequestMsg{}}

	err := messenger.sender.SendMessage(to, msg)
	if err != nil {
		return fmt.Errorf("failed to send order request message to player %s: %w", to, err)
	}

	return nil
}

func (messenger Messenger) SendOrdersReceived(playerOrders map[string][]board.Order) error {
	msg := message{ordersReceivedMsgID: ordersReceivedMsg{PlayerOrders: playerOrders}}

	err := messenger.sender.SendMessageToAll(msg)
	if err != nil {
		return fmt.Errorf("failed to send orders received message: %w", err)
	}

	return nil
}

func (messenger Messenger) SendOrdersConfirmation(player string) error {
	msg := message{ordersConfirmationMsgID: ordersConfirmationMsg{Player: player}}

	err := messenger.sender.SendMessageToAll(msg)
	if err != nil {
		return fmt.Errorf("failed to send orders confirmation message: %w", err)
	}

	return nil
}

func (messenger Messenger) SendSupportRequest(to string, supportingArea string, supportablePlayers []string) error {
	msg := message{
		supportRequestMsgID: supportRequestMsg{SupportingArea: supportingArea, SupportablePlayers: supportablePlayers},
	}

	err := messenger.sender.SendMessage(to, msg)
	if err != nil {
		return fmt.Errorf("failed to send support request message to player %s: %w", to, err)
	}

	return nil
}

func (messenger Messenger) SendBattleResults(battles []board.Battle) error {
	msg := message{battleResultsMsgID: battleResultsMsg{Battles: battles}}

	err := messenger.sender.SendMessageToAll(msg)
	if err != nil {
		return fmt.Errorf("failed to send battle results message: %w", err)
	}

	return nil
}

func (messenger Messenger) SendWinner(winner string) error {
	msg := message{winnerMsgID: winnerMsg{Winner: winner}}

	err := messenger.sender.SendMessageToAll(msg)
	if err != nil {
		return fmt.Errorf("failed to send winner message: %w", err)
	}

	return nil
}
