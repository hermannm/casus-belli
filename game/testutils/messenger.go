package testutils

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

type MockMessenger struct{}

func (MockMessenger) SendBattleResults(battles []gametypes.Battle) error {
	return nil
}

func (MockMessenger) SendSupportRequest(
	toPlayer string, supportingRegion string, embattledRegion string, supportablePlayers []string,
) error {
	return nil
}

func (MockMessenger) ReceiveSupport(
	fromPlayer string, supportingRegion string, embattledRegion string,
) (supportedPlayer string, err error) {
	return "", nil
}
