package lobby

import (
	"errors"
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/wrap"
)

func (player *Player) sendMessage(message Message) error {
	player.lock.Lock()
	defer player.lock.Unlock()

	if err := player.socket.WriteJSON(message); err != nil {
		return wrap.Errorf(
			err, "failed to send message of type '%s' to player %s", message.Type(), player.String(),
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
			message.Type(),
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
			PlayerStatusMessage{Username: player.username, GameID: gameID, ReadyToStartGame: player.readyToStartGame},
		)

		player.lock.RUnlock()
	}

	if err := player.sendMessage(Message{MessageTypeLobbyJoined: LobbyJoinedMessage{
		SelectableGameIDs: lobby.game.PlayerIDs, PlayerStatuses: statuses,
	}}); err != nil {
		return wrap.Errorf(err, "failed to send lobby joined message to player %s", player.String())
	}

	return nil
}

func (lobby *Lobby) SendPlayerStatusMessage(player *Player) error {
	player.lock.RLock()

	statusMsg := PlayerStatusMessage{
		Username: player.username, GameID: nil, ReadyToStartGame: player.readyToStartGame,
	}
	if player.gameID != "" {
		gameID := player.gameID
		statusMsg.GameID = &gameID
	}

	player.lock.RUnlock()

	if err := lobby.sendMessageToAll(Message{MessageTypePlayerStatus: statusMsg}); err != nil {
		return wrap.Error(err, "failed to send player status message")
	}

	return nil
}

func (player *Player) SendError(err error) {
	if err := player.sendMessage(Message{
		MessageTypeError: ErrorMessage{Error: err.Error()},
	}); err != nil {
		log.Println(err)
	}
}

func (lobby *Lobby) SendError(toPlayer string, err error) {
	if err := lobby.sendMessage(toPlayer, Message{
		MessageTypeError: ErrorMessage{Error: err.Error()},
	}); err != nil {
		log.Println(err)
	}
}

func (lobby *Lobby) SendOrderRequest(toPlayer string) error {
	return lobby.sendMessage(toPlayer, Message{
		MessageTypeOrderRequest: OrderRequestMessage{},
	})
}

func (lobby *Lobby) SendOrdersReceived(playerOrders map[string][]gametypes.Order) error {
	return lobby.sendMessageToAll(Message{
		MessageTypeOrdersReceived: OrdersReceivedMessage{PlayerOrders: playerOrders},
	})
}

func (lobby *Lobby) SendOrdersConfirmation(playerWhoSubmittedOrders string) error {
	return lobby.sendMessageToAll(Message{
		MessageTypeOrdersConfirmation: OrdersConfirmationMessage{
			PlayerWhoSubmittedOrders: playerWhoSubmittedOrders,
		},
	})
}

func (lobby *Lobby) SendSupportRequest(
	toPlayer string, supportingRegion string, embattledRegion string, supportablePlayers []string,
) error {
	return lobby.sendMessage(toPlayer, Message{
		MessageTypeSupportRequest: SupportRequestMessage{
			SupportingRegion:   supportingRegion,
			EmbattledRegion:    embattledRegion,
			SupportablePlayers: supportablePlayers,
		},
	})
}

func (lobby *Lobby) SendBattleResults(battles []gametypes.Battle) error {
	return lobby.sendMessageToAll(Message{
		MessageTypeBattleResults: BattleResultsMessage{Battles: battles},
	})
}

func (lobby *Lobby) SendWinner(winner string) error {
	return lobby.sendMessageToAll(Message{
		MessageTypeWinner: WinnerMessage{Winner: winner},
	})
}
