package lobby

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"

	"hermannm.dev/casus-belli/server/game"
)

func (player *Player) sendMessage(message Message) (succeeded bool) {
	player.lock.Lock()
	defer player.lock.Unlock()

	if err := player.socket.WriteJSON(message); err != nil {
		player.log.Error(nil, err, "Failed to send message", "message", message)
		return false
	}

	return true
}

func (lobby *Lobby) sendMessage(to game.PlayerFaction, message Message) (succeeded bool) {
	player, ok := lobby.getPlayer(to)
	if !ok {
		lobby.log.ErrorMessage(
			nil,
			fmt.Sprintf("Tried to send message to unrecognized player faction '%s'", to),
			"message", message,
		)
		return false
	}

	return player.sendMessage(message)
}

func (player *Player) sendPreparedMessage(message *websocket.PreparedMessage) error {
	player.lock.Lock()
	defer player.lock.Unlock()

	return player.socket.WritePreparedMessage(message)
}

func (lobby *Lobby) sendMessageToAll(message Message) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		lobby.log.Error(nil, err, "Failed to serialize message", "message", message)
		return
	}

	preparedMessage, err := websocket.NewPreparedMessage(websocket.TextMessage, messageJSON)
	if err != nil {
		lobby.log.Error(nil, err, "Failed to prepare websocket message", "message", message)
		return
	}

	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	for _, player := range lobby.players {
		if err := player.sendPreparedMessage(preparedMessage); err != nil {
			player.log.Error(nil, err, "Failed to send prepared message", "message", message)
		}
	}
}

func (player *Player) SendLobbyJoinedMessage(lobby *Lobby) {
	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	statuses := make([]PlayerStatusMessage, 0, len(lobby.players)-1)

	for _, otherPlayer := range lobby.players {
		if otherPlayer.username == player.username {
			continue
		}

		otherPlayer.lock.RLock()
		status := PlayerStatusMessage{
			Username:        otherPlayer.username,
			SelectedFaction: otherPlayer.gameFaction,
		}
		otherPlayer.lock.RUnlock()

		statuses = append(statuses, status)
	}

	player.sendMessage(
		Message{
			Tag: MessageTagLobbyJoined,
			Data: LobbyJoinedMessage{
				SelectableFactions: lobby.game.PlayerFactions,
				PlayerStatuses:     statuses,
			},
		},
	)
}

func (lobby *Lobby) SendPlayerStatusMessage(player *Player) {
	player.lock.RLock()
	message := PlayerStatusMessage{
		Username:        player.username,
		SelectedFaction: player.gameFaction,
	}
	player.lock.RUnlock()

	lobby.sendMessageToAll(
		Message{
			Tag:  MessageTagPlayerStatus,
			Data: message,
		},
	)
}

func (player *Player) SendError(err error) {
	player.sendMessage(
		Message{
			Tag:  MessageTagError,
			Data: ErrorMessage{Error: err.Error()},
		},
	)
}

func (lobby *Lobby) SendError(to game.PlayerFaction, err error) {
	lobby.sendMessage(
		to, Message{
			Tag:  MessageTagError,
			Data: ErrorMessage{Error: err.Error()},
		},
	)
}

func (lobby *Lobby) SendGameStarted(board game.Board) {
	lobby.sendMessageToAll(
		Message{
			Tag:  MessageTagGameStarted,
			Data: GameStartedMessage{Board: board},
		},
	)
}

func (lobby *Lobby) SendOrderRequest(to game.PlayerFaction, season game.Season) (succeeded bool) {
	return lobby.sendMessage(
		to, Message{
			Tag:  MessageTagOrderRequest,
			Data: OrderRequestMessage{Season: season},
		},
	)
}

func (lobby *Lobby) SendOrdersReceived(orders map[game.PlayerFaction][]*game.Order) {
	lobby.sendMessageToAll(
		Message{
			Tag:  MessageTagOrdersReceived,
			Data: OrdersReceivedMessage{OrdersByFaction: orders},
		},
	)
}

func (lobby *Lobby) SendOrdersConfirmation(factionThatSubmittedOrders game.PlayerFaction) {
	lobby.sendMessageToAll(
		Message{
			Tag:  MessageTagOrdersConfirmation,
			Data: OrdersConfirmationMessage{FactionThatSubmittedOrders: factionThatSubmittedOrders},
		},
	)
}

func (lobby *Lobby) SendBattleAnnouncement(battle game.Battle) {
	lobby.sendMessageToAll(
		Message{
			Tag:  MessageTagBattleAnnouncement,
			Data: BattleAnnouncementMessage{Battle: battle},
		},
	)
}

func (lobby *Lobby) SendBattleResults(battle game.Battle) {
	lobby.sendMessageToAll(
		Message{
			Tag:  MessageTagBattleResults,
			Data: BattleResultsMessage{Battle: battle},
		},
	)
}

func (lobby *Lobby) SendWinner(winner game.PlayerFaction) {
	lobby.sendMessageToAll(
		Message{
			Tag:  MessageTagWinner,
			Data: WinnerMessage{WinningFaction: winner},
		},
	)
}
