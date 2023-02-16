package testutils

import (
	"hermannm.dev/bfh-server/game/board"
)

type MockMessenger struct{}

func (MockMessenger) SendBattleResults(battles []board.Battle) error {
	return nil
}

func (MockMessenger) SendSupportRequest(to string, supportingRegion string, battlers []string) error {
	return nil
}

func (MockMessenger) ReceiveSupport(from string, fromRegion string) (supportTo string, err error) {
	return "", nil
}
