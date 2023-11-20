package lobby

import (
	"errors"

	"hermannm.dev/bfh-server/game"
	"hermannm.dev/wrap"
)

func (player *Player) sendMessage(message Message) error {
	player.lock.Lock()
	defer player.lock.Unlock()

	if err := player.socket.WriteJSON(message); err != nil {
		return wrap.Errorf(
			err,
			"failed to send message of type '%s' to player '%s'",
			message.Tag,
			player.username,
		)
	}

	return nil
}

func (lobby *Lobby) sendMessage(to game.PlayerFaction, message Message) error {
	player, ok := lobby.getPlayer(to)
	if !ok {
		return wrap.Errorf(
			errors.New("player not found"),
			"failed to send message of type '%s' to player faction '%s'",
			message.Tag,
			to,
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

	statuses := make([]PlayerStatusMessage, 0, len(lobby.players)-1)

	for _, otherPlayer := range lobby.players {
		if otherPlayer.username == player.username {
			continue
		}

		otherPlayer.lock.RLock()

		var faction *game.PlayerFaction
		if otherPlayer.gameFaction != "" {
			faction = &otherPlayer.gameFaction
		}

		statuses = append(
			statuses,
			PlayerStatusMessage{
				Username:        otherPlayer.username,
				SelectedFaction: faction,
			},
		)

		otherPlayer.lock.RUnlock()
	}

	if err := player.sendMessage(Message{
		Tag: MessageTagLobbyJoined,
		Data: LobbyJoinedMessage{
			SelectableFactions: lobby.game.PlayerFactions,
			PlayerStatuses:     statuses,
		},
	}); err != nil {
		return err
	}

	return nil
}

func (lobby *Lobby) SendPlayerStatusMessage(player *Player) error {
	player.lock.RLock()

	statusMsg := PlayerStatusMessage{
		Username:        player.username,
		SelectedFaction: nil,
	}
	if player.gameFaction != "" {
		faction := player.gameFaction
		statusMsg.SelectedFaction = &faction
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
		player.log.Error(err)
	}
}

func (lobby *Lobby) SendError(to game.PlayerFaction, err error) {
	if err := lobby.sendMessage(to, Message{
		Tag:  MessageTagError,
		Data: ErrorMessage{Error: err.Error()},
	}); err != nil {
		lobby.log.Error(err)
	}
}

func (lobby *Lobby) SendGameStarted(board game.Board) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagGameStarted,
		Data: GameStartedMessage{Board: board},
	})
}

func (lobby *Lobby) SendOrderRequest(to game.PlayerFaction, season game.Season) error {
	return lobby.sendMessage(to, Message{
		Tag:  MessageTagOrderRequest,
		Data: OrderRequestMessage{Season: season},
	})
}

func (lobby *Lobby) SendOrdersReceived(orders map[game.PlayerFaction][]game.Order) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagOrdersReceived,
		Data: OrdersReceivedMessage{OrdersByFaction: orders},
	})
}

func (lobby *Lobby) SendOrdersConfirmation(
	factionThatSubmittedOrders game.PlayerFaction,
) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagOrdersConfirmation,
		Data: OrdersConfirmationMessage{FactionThatSubmittedOrders: factionThatSubmittedOrders},
	})
}

func (lobby *Lobby) SendSupportRequest(
	to game.PlayerFaction,
	supporting game.RegionName,
	embattled game.RegionName,
	supportable []game.PlayerFaction,
) error {
	return lobby.sendMessage(to, Message{
		Tag: MessageTagSupportRequest,
		Data: SupportRequestMessage{
			SupportingRegion:    supporting,
			EmbattledRegion:     embattled,
			SupportableFactions: supportable,
		},
	})
}

func (lobby *Lobby) SendBattleResults(battles ...game.Battle) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagBattleResults,
		Data: BattleResultsMessage{Battles: battles},
	})
}

func (lobby *Lobby) SendWinner(winner game.PlayerFaction) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagWinner,
		Data: WinnerMessage{WinningFaction: winner}})
}
