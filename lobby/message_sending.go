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

func (lobby *Lobby) sendMessage(to gametypes.PlayerFaction, message Message) error {
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

	statuses := make([]PlayerStatusMessage, 0, len(lobby.players))

	for _, player := range lobby.players {
		player.lock.RLock()

		var faction *gametypes.PlayerFaction
		if player.gameFaction != "" {
			faction = &player.gameFaction
		}

		statuses = append(
			statuses,
			PlayerStatusMessage{
				Username:         player.username,
				SelectedFaction:  faction,
				ReadyToStartGame: player.readyToStartGame,
			},
		)

		player.lock.RUnlock()
	}

	if err := player.sendMessage(Message{
		Tag: MessageTagLobbyJoined,
		Data: LobbyJoinedMessage{
			SelectableFactions: lobby.game.Factions,
			PlayerStatuses:     statuses,
		},
	}); err != nil {
		return wrap.Errorf(err, "failed to send lobby joined message to player %s", player.String())
	}

	return nil
}

func (lobby *Lobby) SendPlayerStatusMessage(player *Player) error {
	player.lock.RLock()

	statusMsg := PlayerStatusMessage{
		Username:         player.username,
		SelectedFaction:  nil,
		ReadyToStartGame: player.readyToStartGame,
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
		log.Error(err, "")
	}
}

func (lobby *Lobby) SendError(to gametypes.PlayerFaction, err error) {
	if err := lobby.sendMessage(to, Message{
		Tag:  MessageTagError,
		Data: ErrorMessage{Error: err.Error()},
	}); err != nil {
		log.Error(err, "")
	}
}

func (lobby *Lobby) SendOrderRequest(to gametypes.PlayerFaction) error {
	return lobby.sendMessage(to, Message{
		Tag:  MessageTagOrderRequest,
		Data: OrderRequestMessage{},
	})
}

func (lobby *Lobby) SendOrdersReceived(orders map[gametypes.PlayerFaction][]gametypes.Order) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagOrdersReceived,
		Data: OrdersReceivedMessage{OrdersByFaction: orders},
	})
}

func (lobby *Lobby) SendOrdersConfirmation(
	factionThatSubmittedOrders gametypes.PlayerFaction,
) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagOrdersConfirmation,
		Data: OrdersConfirmationMessage{FactionThatSubmittedOrders: factionThatSubmittedOrders},
	})
}

func (lobby *Lobby) SendSupportRequest(
	to gametypes.PlayerFaction,
	supportingRegion string,
	embattledRegion string,
	supportableFactions []gametypes.PlayerFaction,
) error {
	return lobby.sendMessage(to, Message{
		Tag: MessageTagSupportRequest,
		Data: SupportRequestMessage{
			SupportingRegion:    supportingRegion,
			EmbattledRegion:     embattledRegion,
			SupportableFactions: supportableFactions,
		},
	})
}

func (lobby *Lobby) SendBattleResults(battles []gametypes.Battle) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagBattleResults,
		Data: BattleResultsMessage{Battles: battles},
	})
}

func (lobby *Lobby) SendWinner(winner gametypes.PlayerFaction) error {
	return lobby.sendMessageToAll(Message{
		Tag:  MessageTagWinner,
		Data: WinnerMessage{WinningFaction: winner}})
}
