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
	"hermannm.dev/wrap"
)

type LobbyAPI struct {
	router          *http.ServeMux
	lobbyRegistry   *lobby.LobbyRegistry
	availableBoards []boardconfig.BoardInfo
}

func NewLobbyAPI(
	router *http.ServeMux,
	lobbyRegistry *lobby.LobbyRegistry,
	availableBoards []boardconfig.BoardInfo,
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
	lobbyList := api.lobbyRegistry.LobbyInfo()

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
func (api LobbyAPI) JoinLobby(res http.ResponseWriter, req *http.Request) {
	const lobbyNameParam = "lobbyName"
	const usernameParam = "username"

	params, ok := checkParams(req, lobbyNameParam, usernameParam)
	if !ok {
		http.Error(res, "lobby name and username are required to join lobby", http.StatusBadRequest)
		return
	}

	lobbyName := params.Get(lobbyNameParam)
	lobby, ok := api.lobbyRegistry.GetLobby(lobbyName)
	if !ok {
		http.Error(
			res,
			fmt.Sprintf("no lobby found with name '%s'", lobbyName),
			http.StatusNotFound,
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
		fmt.Println(wrap.Error(err, "failed to establish socket connection"))
		http.Error(res, "unable to establish socket connection", http.StatusInternalServerError)
		return
	}

	username := params.Get(usernameParam)

	player, err := lobby.AddPlayer(username, socket)
	if err != nil {
		fmt.Println(wrap.Error(err, "failed to add player"))
		player.SendError(wrap.Error(err, "failed to join game"))
		return
	}

	player.SendLobbyJoinedMessage(lobby)
}

// Endpoint for creating lobbies (for servers with public lobby creation).
// Expects query parameters "lobbyName" and "gameName".
func (api LobbyAPI) CreateLobby(res http.ResponseWriter, req *http.Request) {
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

	if err := api.lobbyRegistry.RegisterLobby(lobby); err != nil {
		err = wrap.Error(err, "failed to register lobby")
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Write([]byte("lobby created"))
}

// Endpoint for showing the list of boards supported by the server.
func (api LobbyAPI) ListBoards(res http.ResponseWriter, req *http.Request) {
	jsonResponse, err := json.Marshal(api.availableBoards)
	if err != nil {
		http.Error(res, "error fetching list of games", http.StatusInternalServerError)
		return
	}

	res.Write(jsonResponse)
}
