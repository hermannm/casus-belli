package lobby

import (
	"errors"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
)

func (player *Player) sendMessage(message Message) error {
	player.lock.Lock()
	defer player.lock.Unlock()

	if err := player.socket.WriteJSON(message); err != nil {
		return wrap.Errorf(
			err,
			"failed to send message of type '%s' to player %s",
			message.Tag,
			player.String(),
		)
	}

	return nil
}

func (lobby *Lobby) sendMessage(toPlayer string, message Message) error {
	player, ok := lobby.getPlayer(toPlayer)
	if !ok {
		return wrap.Errorf(
			errors.New("player not found"),
			"failed to send message of type '%s' to player with game ID '%s'",
			message.Tag,
			toPlayer,
		)
	}

	return player.sendMessage(message)
}

func (lobby *Lobby) sendMessageToAll(message Message) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	var errs []error
	for _, player := range lobby.players {
		if err := player.sendMessage(message); err != nil {
			errs = append(errs, err)
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return wrap.Errors("failed to send message to multiple players", errs...)
	}
}

func (player *Player) SendLobbyJoinedMessage(lobby *Lobby) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	statuses := make([]PlayerStatusMessage, 0, len(lobby.players))

	for _, player := range lobby.players {
		player.lock.RLock()

		var gameID *string
		if player.gameID != "" {
			gameID = &player.gameID
		}

		statuses = append(
			statuses,
			PlayerStatusMessage{
				Username:         player.username,
				GameID:           gameID,
				ReadyToStartGame: player.readyToStartGame,
			},
		)

		player.lock.RUnlock()
	}

	if err := player.sendMessage(Message{
		Tag:  MessageTagLobbyJoined,
		Data: LobbyJoinedMessage{SelectableGameIDs: lobby.game.PlayerIDs, PlayerStatuses: statuses},
	}); err != nil {
		return wrap.Errorf(err, "failed to send lobby joined message to player %s", player.String())
	}

	return nil
}

func (lobby *Lobby) SendPlayerStatusMessage(player *Player) error {
	player.lock.RLock()

	statusMsg := PlayerStatusMessage{
		Username:         player.username,
		GameID:           nil,
		ReadyToStartGame: player.readyToStartGame,
	}
	if player.gameID != "" {
		gameID := player.gameID
		statusMsg.GameID = &gameID
	}

	player.lock.RUnlock()

	if err := lobby.sendMessageToAll(Message{
		Tag:  MessageTagPlayerStatus,
		Data: statusMsg,
	}); err != nil {
		return wrap.Error(err, "failed to send player status message")
	}

	return nil
}

func (player *Player) SendError(err error) {
	if err := player.sendMessage(Message{
		Tag:  MessageTagError,
		Data: ErrorMessage{Error: err.Error()},
	}); err != nil {
		log.Error(err, "")
	}
}

func (lobby *Lobby) SendError(toPlayer string, err error) {
	if err := lobby.sendMessage(toPlayer, Message{
		Tag:  MessageTagError,
		Data: ErrorMessage{Error: err.Error()},
	}); err != nil {
		log.Error(err, "")
	}
}

func (lobby *Lobby) SendOrderRequest(toPlayer string) error {
	return lobby.sendMessage(toPlayer, Message{
		Tag:  MessageTagOrderRequest,
		Data: OrderRequestMessage{},
	})
}

func (lobby *Lobby) SendOrdersReceived(playerOrders map[string][]gametypes.Order) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagOrdersReceived,
		Data: OrdersReceivedMessage{PlayerOrders: playerOrders},
	})
}

func (lobby *Lobby) SendOrdersConfirmation(playerWhoSubmittedOrders string) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagOrdersConfirmation,
		Data: OrdersConfirmationMessage{PlayerWhoSubmittedOrders: playerWhoSubmittedOrders},
	})
}

func (lobby *Lobby) SendSupportRequest(
	toPlayer string, supportingRegion string, embattledRegion string, supportablePlayers []string,
) error {
	return lobby.sendMessage(toPlayer, Message{
		Tag: MessageTagSupportRequest,
		Data: SupportRequestMessage{
			SupportingRegion:   supportingRegion,
			EmbattledRegion:    embattledRegion,
			SupportablePlayers: supportablePlayers,
		},
	})
}

func (lobby *Lobby) SendBattleResults(battles []gametypes.Battle) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagBattleResults,
		Data: BattleResultsMessage{Battles: battles},
	})
}

func (lobby *Lobby) SendWinner(winner string) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagWinner,
		Data: WinnerMessage{Winner: winner}})
}
