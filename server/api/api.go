package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"

	"hermannm.dev/casus-belli/server/game"
	"hermannm.dev/casus-belli/server/lobby"
)

type LobbyAPI struct {
	router          *http.ServeMux
	lobbyRegistry   *lobby.LobbyRegistry
	availableBoards []game.BoardInfo
}

func NewLobbyAPI(
	router *http.ServeMux,
	lobbyRegistry *lobby.LobbyRegistry,
	availableBoards []game.BoardInfo,
) LobbyAPI {
	if router == nil {
		router = http.DefaultServeMux
	}

	api := LobbyAPI{lobbyRegistry: lobbyRegistry, availableBoards: availableBoards, router: router}

	router.HandleFunc("GET /lobbies", api.listLobbies)
	router.HandleFunc("GET /join", api.joinLobby)

	return api
}

func (api LobbyAPI) RegisterLobbyCreationEndpoints() {
	api.router.HandleFunc("POST /create", api.createLobby)
	api.router.HandleFunc("GET /boards", api.listBoards)
}

func (api LobbyAPI) ListenAndServe(address string) error {
	server := &http.Server{
		Addr:              address,
		Handler:           api.router,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		return wrap.Error(err, "server stopped")
	}
	return nil
}

// Endpoint to list available game lobbies.
func (api LobbyAPI) listLobbies(res http.ResponseWriter, _ *http.Request) {
	sendJSON(res, api.lobbyRegistry.ListLobbies())
}

// Endpoint for a player to join a lobby.
// Expects query parameters "lobbyName" and "username".
func (api LobbyAPI) joinLobby(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	query := req.URL.Query()

	lobbyName, err := getQueryParam(query, "lobbyName")
	if err != nil {
		sendClientErrorWithHeader(res, err)
		return
	}

	username, err := getQueryParam(query, "username")
	if err != nil {
		sendClientErrorWithHeader(res, err)
		return
	}

	gameLobby, ok := api.lobbyRegistry.GetLobby(lobbyName)
	if !ok {
		sendClientErrorWithHeader(res, fmt.Errorf("no lobby found with name '%s'", lobbyName))
		return
	}

	//nolint:exhaustruct
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Accepts all origins for now, in order to enable clients from other networks.
		CheckOrigin: func(*http.Request) bool { return true },
	}

	socket, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		err = wrap.Error(err, "failed to establish socket connection")
		sendServerErrorWithHeader(res, err)
		gameLobby.Logger().Error(ctx, err, "player", username)
		return
	}

	player, err := gameLobby.AddPlayer(username, socket)
	if err != nil {
		gameLobby.Logger().Error(ctx, err, "failed to add player", "player", username)
		_ = socket.WriteJSON(
			lobby.Message{
				Tag:  lobby.MessageTagError,
				Data: lobby.ErrorMessage{Error: wrap.Error(err, "failed to join game").Error()},
			},
		)
		_ = socket.Close()
		return
	}

	player.SendLobbyJoinedMessage(gameLobby)
	gameLobby.SendPlayerStatusMessage(player)
}

// Endpoint for creating lobbies (for servers with public lobby creation enabled).
// Expects query parameters "lobbyName" and "boardID".
func (api LobbyAPI) createLobby(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	query := req.URL.Query()

	lobbyName, err := getQueryParam(query, "lobbyName")
	if err != nil {
		sendClientError(res, err)
		return
	}

	boardID, err := getQueryParam(query, "boardID")
	if err != nil {
		sendClientError(res, err)
		return
	}

	if err := api.lobbyRegistry.CreateLobby(lobbyName, boardID, false, nil); err != nil {
		err = wrap.Error(err, "failed to create lobby")
		sendServerError(res, err)
		log.Error(ctx, err, "")
		return
	}

	res.WriteHeader(http.StatusCreated)
}

// Endpoint for showing the list of boards supported by the server.
func (api LobbyAPI) listBoards(res http.ResponseWriter, _ *http.Request) {
	sendJSON(res, api.availableBoards)
}
