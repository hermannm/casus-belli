package api

import (
	"net/http"
)

type Lobby struct {
	ID         string
	MinPlayers int
	MaxPlayers int
	Players    map[string]string
}

func CreateLobby(id string, minPlayers int, playerIDs []string) {
	lobby := Lobby{
		ID:         id,
		MinPlayers: minPlayers,
		MaxPlayers: len(playerIDs),
		Players:    make(map[string]string, len(playerIDs)),
	}
	for _, playerID := range playerIDs {
		lobby.Players[playerID] = ""
	}

	http.HandleFunc("/:id/join", receivePlayer)
}

func receivePlayer(writer http.ResponseWriter, req *http.Request) {

}
