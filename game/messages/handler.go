package messages

import (
	"fmt"
	"log"

	"hermannm.dev/bfh-server/game/board"
)

type Handler struct {
	sender    Sender
	receivers map[string]Receiver
}

func NewHandler(s Sender) Handler {
	rs := make(map[string]Receiver)
	return Handler{
		sender:    s,
		receivers: rs,
	}
}

func (h Handler) SendError(to string, errMsg string) {
	err := h.sender.SendMessage(to, Error{Type: MsgError, Error: errMsg})
	if err != nil {
		log.Println(fmt.Errorf("failed to send error message to player %s: %w", to, err))
	}
}

func (h Handler) SendOrderRequest(to string) error {
	err := h.sender.SendMessage(to, OrderRequest{Type: MsgOrderRequest})
	if err != nil {
		return fmt.Errorf("failed to send order request message to player %s: %w", to, err)
	}
	return nil
}

func (h Handler) SendSupportRequest(to string, supportingArea string, battlers []string) error {
	err := h.sender.SendMessage(to, SupportRequest{
		Type:           MsgSupportRequest,
		SupportingArea: supportingArea,
		Battlers:       battlers,
	})
	if err != nil {
		return fmt.Errorf("failed to send support request message to player %s: %w", to, err)
	}
	return nil
}

func (h Handler) SendBattleResult(battle board.Battle) error {
	err := h.sender.SendMessageToAll(BattleResult{Type: MsgBattleResult, Battle: battle})
	if err != nil {
		return fmt.Errorf("failed to send battle result message: %w", err)
	}
	return nil
}

func (h Handler) SendWinner(winner string) error {
	err := h.sender.SendMessageToAll(Winner{Type: MsgWinner, Winner: winner})
	if err != nil {
		return fmt.Errorf("failed to send winner message: %w", err)
	}
	return nil
}

func (h Handler) ReceiveOrders(from string) ([]board.Order, error) {
	receiver, ok := h.receivers[from]
	if !ok {
		return nil, fmt.Errorf("failed to get order message from player %s: receiver not found", from)
	}

	orders := <-receiver.Orders
	return orders.Orders, nil
}

func (h Handler) ReceiveSupport(from string, supportingArea string) (supportTo string, err error) {
	receiver, ok := h.receivers[from]
	if !ok {
		return "", fmt.Errorf("failed to get support message from player %s: receiver not found", from)
	}

	supportChan, ok := receiver.Support[supportingArea]
	if !ok {
		return "", fmt.Errorf(
			"failed to get support message from player %s: support receiver uninitialized for area %s",
			from,
			supportingArea,
		)
	}

	support := <-supportChan
	return support.Player, nil
}
