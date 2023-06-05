package api

import (
	"net/http"

	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/bfh-server/lobby"
)

func RegisterEndpoints(mux *http.ServeMux, lobbyRegistry *lobby.LobbyRegistry) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	mux.Handle("/lobbies", LobbyListHandler{lobbyRegistry})
	mux.Handle("/join", JoinLobbyHandler{lobbyRegistry})
}

func RegisterLobbyCreationEndpoints(
	mux *http.ServeMux, lobbyRegistry *lobby.LobbyRegistry, availableBoards []boardconfig.BoardInfo,
) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	mux.Handle("/create", CreateLobbyHandler{lobbyRegistry})
	mux.Handle("/boards", BoardListHandler{availableBoards})
}
