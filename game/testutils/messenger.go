package testutils

import (
	"hermannm.dev/bfh-server/game/board"
)

type MockMessenger struct{}

func (MockMessenger) SendBattleResults(battles []board.Battle) error {
	return nil
}

func (MockMessenger) SendSupportRequest(to string, supportingArea string, battlers []string) error {
	return nil
}

func (MockMessenger) ReceiveSupport(from string, fromArea string) (supportTo string, err error) {
	return "", nil
}
