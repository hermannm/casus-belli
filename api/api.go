package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/bfh-server/lobby"
)

// Endpoint to list available game lobbies.
type LobbyListHandler struct {
	lobbyRegistry *lobby.LobbyRegistry
}

func (handler LobbyListHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	lobbyList := handler.lobbyRegistry.LobbyInfo()

	lobbyListJSON, err := json.Marshal(lobbyList)
	if err != nil {
		http.Error(
			res,
			"error in reading lobby fetching lobby list",
			http.StatusInternalServerError,
		)
		return
	}

	res.Write(lobbyListJSON)
}

// Endpoint for a player to join a lobby.
// Expects query parameters "lobbyName" and "username".
type JoinLobbyHandler struct {
	lobbyRegistry *lobby.LobbyRegistry
}

func (handler JoinLobbyHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	const lobbyNameParam = "lobbyName"
	const usernameParam = "username"

	params, ok := checkParams(req, lobbyNameParam, usernameParam)
	if !ok {
		http.Error(
			res, "lobby name and username are required to join lobby", http.StatusBadRequest,
		)
		return
	}

	lobbyName := params.Get(lobbyNameParam)
	lobby, ok := handler.lobbyRegistry.GetLobby(lobbyName)
	if !ok {
		http.Error(
			res, fmt.Sprintf("no lobby found with name '%s'", lobbyName), http.StatusNotFound,
		)
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
		log.Println(fmt.Errorf("failed to establish socket connection: %w", err))
		http.Error(res, "unable to establish socket connection", http.StatusInternalServerError)
		return
	}

	username := params.Get(usernameParam)

	player, err := lobby.AddPlayer(username, socket)
	if err != nil {
		log.Println(fmt.Errorf("failed to add player: %w", err))
		player.SendError(fmt.Errorf("failed to join game: %w", err))
		return
	}

	player.SendLobbyJoinedMessage(lobby)
}

// Endpoint for creating lobbies (for servers with public lobby creation).
// Expects query parameters "lobbyName" and "gameName".
type CreateLobbyHandler struct {
	lobbyRegistry *lobby.LobbyRegistry
}

func (handler CreateLobbyHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	const lobbyNameParam = "lobbyName"
	const gameNameParam = "gameName"

	params, ok := checkParams(req, lobbyNameParam, gameNameParam)
	if !ok {
		http.Error(res, "insufficient query parameters", http.StatusBadRequest)
		return
	}

	lobbyName, err := url.QueryUnescape(params.Get(lobbyNameParam))
	if err != nil {
		http.Error(res, "invalid lobby name provided", http.StatusBadRequest)
		return
	}

	gameName, err := url.QueryUnescape(params.Get(gameNameParam))
	if err != nil {
		http.Error(res, "invalid game title provided", http.StatusBadRequest)
	}

	lobby, err := lobby.New(lobbyName, gameName, game.DefaultOptions())
	if err != nil {
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := handler.lobbyRegistry.RegisterLobby(lobby); err != nil {
		err = fmt.Errorf("failed to register lobby: %w", err)
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Write([]byte("lobby created"))
}

// Endpoint for showing the list of boards supported by the server.
type BoardListHandler struct {
	availableBoards []boardconfig.BoardInfo
}

func (handler BoardListHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	jsonResponse, err := json.Marshal(handler.availableBoards)
	if err != nil {
		http.Error(res, "error fetching list of games", http.StatusInternalServerError)
		return
	}

	res.Write(jsonResponse)
}
