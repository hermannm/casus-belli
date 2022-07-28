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

func (h Handler) SendError(to string, errMsg string) {
	err := h.sender.SendMessage(to, errorMsg{Type: msgError, Error: errMsg})
	if err != nil {
		log.Println(fmt.Errorf("failed to send error message to player %s: %w", to, err))
	}
}

func (h Handler) SendOrderRequest(to string) error {
	err := h.sender.SendMessage(to, orderRequestMsg{Type: msgOrderRequest})
	if err != nil {
		return fmt.Errorf("failed to send order request message to player %s: %w", to, err)
	}
	return nil
}

func (h Handler) SendSupportRequest(to string, supportingArea string, battlers []string) error {
	err := h.sender.SendMessage(to, supportRequestMsg{
		Type:           msgSupportRequest,
		SupportingArea: supportingArea,
		Battlers:       battlers,
	})
	if err != nil {
		return fmt.Errorf("failed to send support request message to player %s: %w", to, err)
	}
	return nil
}

func (h Handler) SendBattleResult(battle board.Battle) error {
	err := h.sender.SendMessageToAll(battleResultMsg{Type: msgBattleResult, Battle: battle})
	if err != nil {
		return fmt.Errorf("failed to send battle result message: %w", err)
	}
	return nil
}

func (h Handler) SendWinner(winner string) error {
	err := h.sender.SendMessageToAll(winnerMsg{Type: msgWinner, Winner: winner})
	if err != nil {
		return fmt.Errorf("failed to send winner message: %w", err)
	}
	return nil
}
