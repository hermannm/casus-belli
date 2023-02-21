package lobby

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

// Registers handlers for the lobby API endpoints on the given ServeMux.
// If nil is passed as the ServeMux, the default http ServeMux is used.
func RegisterEndpoints(mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	// Endpoint for clients to join a given lobby.
	mux.HandleFunc("/join", joinLobby)

	// Endpoint for clients to view info about all lobbies on the server.
	mux.HandleFunc("/lobbies", lobbyList)
}

// Registers handler for public lobby creation endpoint on the given ServeMux.
// If nil is passed as the ServeMux, the default http ServeMux is used.
// The endpoint expects a parameter corresponding to a key in the game constructor map
// in order to know which type of game to create.
func RegisterLobbyCreationEndpoints(mux *http.ServeMux, games map[string]GameConstructor) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	// Endpoint for clients to create their own lobbies if the server is set to enable that.
	mux.Handle("/create", createLobbyHandler{games: games})

	gameTitles := make([]string, 0, len(games))
	for key := range games {
		gameTitles = append(gameTitles, key)
	}

	// Endpoint for clients to view a list of possible games for which they can create lobbies.
	mux.Handle("/games", gameListHandler{gameTitles: gameTitles})
}

// Handler for returning information about all available lobbies.
func lobbyList(res http.ResponseWriter, req *http.Request) {
	lobbyList := lobbyRegistry.lobbyInfo()

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

// Handler for a player to join a lobby.
// Expects query parameters "lobbyName" and "username" (name the user wants to join with).
func joinLobby(res http.ResponseWriter, req *http.Request) {
	const lobbyParam = "lobbyName"
	const usernameParam = "username"

	params, ok := checkParams(req, lobbyParam, usernameParam)
	if !ok {
		http.Error(
			res,
			"lobby name and username are required to join lobby",
			http.StatusBadRequest,
		)
		return
	}

	lobbyName := params.Get(lobbyParam)
	lobby, ok := lobbyRegistry.getLobby(lobbyName)
	if !ok {
		http.Error(
			res,
			fmt.Sprintf("no lobby found with name \"%s\"", lobbyName),
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
		log.Println(fmt.Errorf("failed to establish socket connection: %w", err))
		http.Error(res, "unable to establish socket connection", http.StatusInternalServerError)
		return
	}

	username := params.Get(usernameParam)

	player, err := lobby.addPlayer(username, socket)
	if err != nil {
		log.Println(fmt.Errorf("failed to add player: %w", err))
		player.sendErr("failed to join game")
		return
	}

	player.sendLobbyJoinedMsg(lobby)
}

// Handler for creating lobbies (for servers with public lobby creation).
// Expects query parameters "lobbyName" and "gameName".
type createLobbyHandler struct {
	games map[string]GameConstructor
}

func (handler createLobbyHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	const lobbyNameParam = "lobbyName"
	const gameNameParam = "gameName"

	params, ok := checkParams(req, lobbyNameParam, gameNameParam)
	if !ok {
		http.Error(res, "insufficient query parameters", http.StatusBadRequest)
		return
	}

	lobbyName, err := url.QueryUnescape(params.Get(lobbyNameParam))
	if err != nil {
		http.Error(res, "invalid lobby ID provided", http.StatusBadRequest)
		return
	}

	gameTitle, err := url.QueryUnescape(params.Get(gameNameParam))
	if err != nil {
		http.Error(res, "invalid game title provided", http.StatusBadRequest)
	}

	gameConstructor, ok := handler.games[gameTitle]
	if !ok {
		http.Error(res, "invalid game title provided", http.StatusBadRequest)
		return
	}

	_, err = New(lobbyName, gameConstructor)
	if err != nil {
		http.Error(res, "error creating lobby", http.StatusInternalServerError)
		return
	}

	res.Write([]byte("lobby created"))
}

// Handler for showing the list of games supported by the server.
type gameListHandler struct {
	gameTitles []string
}

func (handler gameListHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	jsonResponse, err := json.Marshal(handler.gameTitles)
	if err != nil {
		http.Error(res, "error fetching list of games", http.StatusInternalServerError)
		return
	}

	res.Write(jsonResponse)
}
