package messages

import (
	"fmt"
	"log"

	"hermannm.dev/bfh-server/game/board"
)

type Sender interface {
	SendMessage(to string, msg any) error
	SendMessageToAll(msg any) error
}

func (handler Handler) SendError(to string, errMsg string) {
	err := handler.sender.SendMessage(to, errorMsg{Type: msgError, Error: errMsg})
	if err != nil {
		log.Println(fmt.Errorf("failed to send error message to player %s: %w", to, err))
	}
}

func (handler Handler) SendOrderRequest(to string) error {
	err := handler.sender.SendMessage(to, orderRequestMsg{Type: msgOrderRequest})
	if err != nil {
		return fmt.Errorf("failed to send order request message to player %s: %w", to, err)
	}
	return nil
}

func (handler Handler) SendOrdersReceived(playerOrders map[string][]board.Order) error {
	err := handler.sender.SendMessageToAll(ordersReceivedMsg{Type: msgOrdersReceived, Orders: playerOrders})
	if err != nil {
		return fmt.Errorf("failed to send orders received message: %w", err)
	}
	return nil
}

func (handler Handler) SendOrdersConfirmation(player string) error {
	err := handler.sender.SendMessageToAll(ordersConfirmationMsg{Type: msgOrdersConfirmation, Player: player})
	if err != nil {
		return fmt.Errorf("failed to send orders confirmation message: %w", err)
	}
	return nil
}

func (handler Handler) SendSupportRequest(to string, supportingArea string, battlers []string) error {
	err := handler.sender.SendMessage(to, supportRequestMsg{
		Type:           msgSupportRequest,
		SupportingArea: supportingArea,
		Battlers:       battlers,
	})
	if err != nil {
		return fmt.Errorf("failed to send support request message to player %s: %w", to, err)
	}
	return nil
}

func (handler Handler) SendBattleResult(battle board.Battle) error {
	err := handler.sender.SendMessageToAll(battleResultMsg{Type: msgBattleResult, Battle: battle})
	if err != nil {
		return fmt.Errorf("failed to send battle result message: %w", err)
	}
	return nil
}

func (handler Handler) SendWinner(winner string) error {
	err := handler.sender.SendMessageToAll(winnerMsg{Type: msgWinner, Winner: winner})
	if err != nil {
		return fmt.Errorf("failed to send winner message: %w", err)
	}
	return nil
}
