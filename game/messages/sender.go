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

func (messenger Messenger) SendError(to string, errMsg string) {
	err := messenger.sender.SendMessage(to, errorMsg{Type: msgError, Error: errMsg})
	if err != nil {
		log.Println(fmt.Errorf("failed to send error message to player %s: %w", to, err))
	}
}

func (messenger Messenger) SendOrderRequest(to string) error {
	err := messenger.sender.SendMessage(to, orderRequestMsg{Type: msgOrderRequest})
	if err != nil {
		return fmt.Errorf("failed to send order request message to player %s: %w", to, err)
	}
	return nil
}

func (messenger Messenger) SendOrdersReceived(playerOrders map[string][]board.Order) error {
	err := messenger.sender.SendMessageToAll(ordersReceivedMsg{Type: msgOrdersReceived, Orders: playerOrders})
	if err != nil {
		return fmt.Errorf("failed to send orders received message: %w", err)
	}
	return nil
}

func (messenger Messenger) SendOrdersConfirmation(player string) error {
	err := messenger.sender.SendMessageToAll(ordersConfirmationMsg{Type: msgOrdersConfirmation, Player: player})
	if err != nil {
		return fmt.Errorf("failed to send orders confirmation message: %w", err)
	}
	return nil
}

func (messenger Messenger) SendSupportRequest(to string, supportingArea string, battlers []string) error {
	err := messenger.sender.SendMessage(to, supportRequestMsg{
		Type:           msgSupportRequest,
		SupportingArea: supportingArea,
		Battlers:       battlers,
	})
	if err != nil {
		return fmt.Errorf("failed to send support request message to player %s: %w", to, err)
	}
	return nil
}

func (messenger Messenger) SendBattleResults(battles []board.Battle) error {
	err := messenger.sender.SendMessageToAll(battleResultsMsg{Type: msgBattleResults, Battles: battles})
	if err != nil {
		return fmt.Errorf("failed to send battle results message: %w", err)
	}
	return nil
}

func (messenger Messenger) SendWinner(winner string) error {
	err := messenger.sender.SendMessageToAll(winnerMsg{Type: msgWinner, Winner: winner})
	if err != nil {
		return fmt.Errorf("failed to send winner message: %w", err)
	}
	return nil
}
