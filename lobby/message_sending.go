package lobby

import (
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/gametypes"
)

func (player *Player) sendMessage(message Message) error {
	player.lock.Lock()
	defer player.lock.Unlock()

	if err := player.socket.WriteJSON(message); err != nil {
		return fmt.Errorf(
			"failed to send message of type '%s' to player '%s': %w",
			message.Type(), player.String(), err,
		)
	}

	return nil
}

func (lobby *Lobby) sendMessage(toPlayer string, message Message) error {
	player, ok := lobby.getPlayer(toPlayer)
	if !ok {
		return fmt.Errorf(
			"failed to send message of type '%s' to player with game ID '%s': player not found",
			message.Type(), toPlayer,
		)
	}

	return player.sendMessage(message)
}

// Marshals the given message to JSON and sends it to all connected players.
// Returns an error if it failed to marshal or send to at least one of the players.
func (lobby *Lobby) sendMessageToAll(message Message) error {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	var errs []error
	for _, player := range lobby.players {
		err := player.sendMessage(message)
		if err != nil {
			errs = append(errs, err)
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return fmt.Errorf("failed to send message to multiple players:\n%w", errors.Join(errs...))
	}
}

// Gives an overview of other players to a player who has just joined a lobby.
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

	err := player.sendMessage(Message{messageTypeLobbyJoined: LobbyJoinedMessage{
		SelectableGameIDs: lobby.game.PlayerIDs, PlayerStatuses: statuses,
	}})
	if err != nil {
		return fmt.Errorf("failed to send lobby joined message to player %s: %w", player.String(), err)
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

	if err := lobby.sendMessageToAll(Message{messageTypePlayerStatus: statusMsg}); err != nil {
		return fmt.Errorf("failed to send player status message: %w", err)
	}

	return nil
}

func (player *Player) SendError(err error) error {
	return player.sendMessage(Message{
		messageTypeError: ErrorMessage{Error: err.Error()},
	})
}

func (lobby *Lobby) SendError(toPlayer string, err error) error {
	return lobby.sendMessage(toPlayer, Message{
		messageTypeError: ErrorMessage{Error: err.Error()},
	})
}

func (lobby *Lobby) SendOrderRequest(toPlayer string) error {
	return lobby.sendMessage(toPlayer, Message{
		messageTypeOrderRequest: OrderRequestMessage{},
	})
}

func (lobby *Lobby) SendOrdersReceived(playerOrders map[string][]gametypes.Order) error {
	return lobby.sendMessageToAll(Message{
		messageTypeOrdersReceived: OrdersReceivedMessage{PlayerOrders: playerOrders},
	})
}

func (lobby *Lobby) SendOrdersConfirmation(playerWhoSubmittedOrders string) error {
	return lobby.sendMessageToAll(Message{
		messageTypeOrdersConfirmation: OrdersConfirmationMessage{
			PlayerWhoSubmittedOrders: playerWhoSubmittedOrders,
		},
	})
}

func (lobby *Lobby) SendSupportRequest(
	toPlayer string, supportingRegion string, embattledRegion string, supportablePlayers []string,
) error {
	return lobby.sendMessage(toPlayer, Message{
		messageTypeSupportRequest: SupportRequestMessage{
			SupportingRegion:   supportingRegion,
			EmbattledRegion:    embattledRegion,
			SupportablePlayers: supportablePlayers,
		},
	})
}

func (lobby *Lobby) SendBattleResults(battles []gametypes.Battle) error {
	return lobby.sendMessageToAll(Message{
		messageTypeBattleResults: BattleResultsMessage{Battles: battles},
	})
}

func (lobby *Lobby) SendWinner(winner string) error {
	return lobby.sendMessageToAll(Message{
		messageTypeWinner: WinnerMessage{Winner: winner},
	})
}
