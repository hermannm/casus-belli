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
	statuses := make([]lobbyPlayerStatus, 0)

	lobby.lock.RLock()
	defer lobby.lock.RUnlock()

	for _, player := range lobby.players {
		player.lock.RLock()

		var gameID *string
		if player.gameID != "" {
			gameID = &player.gameID
		}

		statuses = append(statuses, lobbyPlayerStatus{Username: player.username, GameID: gameID, Ready: player.ready})

		player.lock.RUnlock()
	}

	err := player.send(message{lobbyJoinedMsgID: lobbyJoinedMsg{PlayerStatuses: statuses}})
	if err != nil {
		return fmt.Errorf("failed to send lobby joined message to player %s: %w", player.String(), err)
	}
	return nil
}
