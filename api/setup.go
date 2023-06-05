package api

import (
	"net/http"

	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/bfh-server/lobby"
)

// Registers handlers for the lobby API endpoints on the given ServeMux.
// If nil is passed as the ServeMux, the default http ServeMux is used.
func RegisterEndpoints(mux *http.ServeMux, lobbyRegistry *lobby.LobbyRegistry) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	mux.Handle("/lobbies", LobbyListHandler{lobbyRegistry})
	mux.Handle("/join", JoinLobbyHandler{lobbyRegistry})
}

// Registers handler for public lobby creation endpoint on the given ServeMux.
// If nil is passed as the ServeMux, the default http ServeMux is used.
// The endpoint expects a parameter corresponding to a key in the game constructor map
// in order to know which type of game to create.
func RegisterLobbyCreationEndpoints(
	mux *http.ServeMux, lobbyRegistry *lobby.LobbyRegistry, availableBoards []boardconfig.BoardInfo,
) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	mux.Handle("/create", CreateLobbyHandler{lobbyRegistry})
	mux.Handle("/boards", BoardListHandler{availableBoards})
}
