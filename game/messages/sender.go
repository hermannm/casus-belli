package messages

import (
	"fmt"
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
)

type Sender interface {
	SendMessage(toPlayer string, msg map[string]any) error
	SendMessageToAll(msg map[string]any) error
}

func (messenger Messenger) SendError(toPlayer string, errMsg string) {
	msg := message{errorMsgID: errorMsg{Error: errMsg}}

	err := messenger.sender.SendMessage(toPlayer, msg)
	if err != nil {
		log.Println(fmt.Errorf("failed to send error message to player %s: %w", toPlayer, err))
	}
}

func (messenger Messenger) SendOrderRequest(toPlayer string) error {
	msg := message{orderRequestMsgID: orderRequestMsg{}}

	err := messenger.sender.SendMessage(toPlayer, msg)
	if err != nil {
		return fmt.Errorf("failed to send order request message to player %s: %w", toPlayer, err)
	}

	return nil
}

func (messenger Messenger) SendOrdersReceived(playerOrders map[string][]gametypes.Order) error {
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

func (messenger Messenger) SendSupportRequest(
	toPlayer string,
	supportingRegion string,
	supportablePlayers []string,
) error {
	msg := message{supportRequestMsgID: supportRequestMsg{
		SupportingRegion: supportingRegion, SupportablePlayers: supportablePlayers},
	}

	err := messenger.sender.SendMessage(toPlayer, msg)
	if err != nil {
		return fmt.Errorf("failed to send support request message to player %s: %w", toPlayer, err)
	}

	return nil
}

func (messenger Messenger) SendBattleResults(battles []gametypes.Battle) error {
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
