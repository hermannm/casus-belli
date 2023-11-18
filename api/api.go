package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/bfh-server/lobby"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
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

	router.HandleFunc("/lobbies", api.ListLobbies)
	router.HandleFunc("/join", api.JoinLobby)

	return api
}

func (api LobbyAPI) RegisterLobbyCreationEndpoints() {
	api.router.HandleFunc("/create", api.CreateLobby)
	api.router.HandleFunc("/boards", api.ListBoards)
}

func (api LobbyAPI) ListenAndServe(address string) error {
	if err := http.ListenAndServe(address, api.router); err != nil {
		return wrap.Error(err, "server stopped")
	}
	return nil
}

// Endpoint to list available game lobbies.
func (api LobbyAPI) ListLobbies(res http.ResponseWriter, req *http.Request) {
	sendJSON(res, api.lobbyRegistry.ListLobbies())
}

// Endpoint for a player to join a lobby.
// Expects query parameters "lobbyName" and "username".
func (api LobbyAPI) JoinLobby(res http.ResponseWriter, req *http.Request) {
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

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Accepts all origins for now, in order to enable clients from other networks.
		CheckOrigin: func(*http.Request) bool { return true },
	}

	socket, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		err := wrap.Error(err, "failed to establish socket connection")
		sendServerErrorWithHeader(res, err)
		gameLobby.Logger().Error(err, slog.String("player", username))
		return
	}

	player, err := gameLobby.AddPlayer(username, socket)
	if err != nil {
		gameLobby.Logger().ErrorCause(err, "failed to add player", slog.String("player", username))
		socket.WriteJSON(lobby.Message{
			Tag:  lobby.MessageTagError,
			Data: lobby.ErrorMessage{Error: wrap.Error(err, "failed to join game").Error()},
		})
		socket.Close()
		return
	}

	player.SendLobbyJoinedMessage(gameLobby)
}

// Endpoint for creating lobbies (for servers with public lobby creation enabled).
// Expects query parameters "lobbyName" and "boardID".
func (api LobbyAPI) CreateLobby(res http.ResponseWriter, req *http.Request) {
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

	if err := api.lobbyRegistry.CreateLobby(lobbyName, boardID, false); err != nil {
		err = wrap.Error(err, "failed to create lobby")
		sendServerError(res, err)
		log.Error(err)
		return
	}

	res.Write([]byte("lobby created"))
}

// Endpoint for showing the list of boards supported by the server.
func (api LobbyAPI) ListBoards(res http.ResponseWriter, req *http.Request) {
	sendJSON(res, api.availableBoards)
}
