package lobby

import (
	"fmt"
)

// Sends the given error message to the player.
func (player *Player) sendErr(errMsg string) {
	player.send(message{errorMsgID: errorMsg{Error: errMsg}})
}

// Gives an overview of other players to a player who has just joined a lobby.
func (player *Player) sendLobbyJoinedMsg(lobby *Lobby) error {
	statuses := make([]playerStatusMsg, 0)

	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	for _, player := range lobby.players {
		player.lock.RLock()

		var gameID *string
		if player.gameID != "" {
			gameID = &player.gameID
		}

		statuses = append(
			statuses,
			playerStatusMsg{Username: player.username, GameID: gameID, Ready: player.ready},
		)

		player.lock.RUnlock()
	}

	gameIDs := lobby.game.PlayerIDs()

	err := player.send(message{lobbyJoinedMsgID: lobbyJoinedMsg{GameIDs: gameIDs, PlayerStatuses: statuses}})
	if err != nil {
		return fmt.Errorf("failed to send lobby joined message to player %s: %w", player.String(), err)
	}
	return nil
}

func (lobby *Lobby) sendPlayerStatusMsg(player *Player) error {
	player.lock.RLock()

	statusMsg := playerStatusMsg{
		Username: player.username,
		GameID:   nil,
		Ready:    player.ready,
	}
	if player.gameID != "" {
		gameID := player.gameID
		statusMsg.GameID = &gameID
	}

	player.lock.RUnlock()

	err := lobby.SendMessageToAll(message{playerStatusMsgID: statusMsg})
	if err != nil {
		return fmt.Errorf("failed to send player status message: %w", err)
	}

	return nil
}
