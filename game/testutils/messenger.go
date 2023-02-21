package testutils

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

type MockMessenger struct{}

func (MockMessenger) SendBattleResults(battles []gametypes.Battle) error {
	return nil
}

func (MockMessenger) SendSupportRequest(
	toPlayer string, supportingRegion string, battlers []string,
) error {
	return nil
}

func (MockMessenger) ReceiveSupport(
	fromPlayer string, fromRegion string,
) (supportedPlayer string, err error) {
	return "", nil
}
