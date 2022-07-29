package lobby

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

// Registers handlers for the lobby API endpoints on the given ServeMux.
// If nil is passed as the ServeMux, the default http ServeMux is used.
func RegisterEndpoints(mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	// Endpoint for clients to join a given lobby.
	// Takes query parameters "lobby" (name of the lobby) and "player" (the player ID that the client wants to claim).
	mux.HandleFunc("/join", addPlayer)

	// Endpoint for clients to view info about a single lobby.
	// Takes query parameter "lobby" (name of the lobby).
	mux.HandleFunc("/info", getLobby)

	// Endpoint for clients to view info about all lobbies on the server.
	mux.HandleFunc("/all", getLobbies)
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
	// Takes query parameters "id" (unique name of the lobby) and "playerIDs".
	mux.Handle("/new", lobbyCreationHandler{games: games})

	gameTitles := make([]string, len(games))
	for key := range games {
		gameTitles = append(gameTitles, key)
	}

	// Endpoint for clients to view a list of possible games for which they can create lobbies.
	mux.Handle("/games", gameListHandler{gameTitles: gameTitles})
}

// Checks the given request for the existence of the provided parameter keys.
// If all exist, returns the parameters, otherwise returns ok = false.
func checkParams(req *http.Request, keys ...string) (params url.Values, ok bool) {
	params = req.URL.Query()

	for _, key := range keys {
		if !params.Has(key) {
			return nil, false
		}
	}

	return params, true
}

// Utility type for responding to requests for lobby info.
type lobbyInfo struct {
	ID                 string          `json:"id"`
	AvailablePlayerIDs map[string]bool `json:"availablePlayerIDs"`
}

// Handler for returning information about a given lobby.
func getLobby(res http.ResponseWriter, req *http.Request) {
	lobby, err := findLobby(req)
	if err != nil {
		http.Error(res, "could not fetch lobby", http.StatusNotFound)
		return
	}

	info, err := json.Marshal(lobbyInfo{
		ID:                 lobby.id,
		AvailablePlayerIDs: lobby.availablePlayerIDs(),
	})
	if err != nil {
		http.Error(res, "could not serialize lobby", http.StatusInternalServerError)
		return
	}

	res.Write(info)
}

// Handler for returning information about all available lobbies.
func getLobbies(res http.ResponseWriter, req *http.Request) {
	lobbyInfoList := make([]lobbyInfo, 0, len(lobbies))

	for _, lobby := range lobbies {
		lobbyInfoList = append(lobbyInfoList, lobbyInfo{
			ID:                 lobby.id,
			AvailablePlayerIDs: lobby.availablePlayerIDs(),
		})
	}

	info, err := json.Marshal(lobbyInfoList)
	if err != nil {
		http.Error(res, "error in reading lobby fetching lobby list", http.StatusInternalServerError)
		return
	}

	res.Write(info)
}

// Handler for adding a player to a lobby.
func addPlayer(res http.ResponseWriter, req *http.Request) {
	lobby, err := findLobby(req)
	if err != nil {
		http.Error(res, "could not find lobby", http.StatusNotFound)
	}

	params, ok := checkParams(req, "player")
	if !ok {
		http.Error(res, "must select player ID", http.StatusBadRequest)
		return
	}

	playerID := params.Get("player")
	lobby.lock.Lock()
	player, ok := lobby.players[playerID]
	if !ok {
		http.Error(res, "invalid player ID", http.StatusBadRequest)
		lobby.lock.Unlock()
		return
	}
	if player.isActive() {
		http.Error(res, "player ID already taken", http.StatusConflict)
		lobby.lock.Unlock()
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
		log.Println(err)
		http.Error(res, "unable to establish socket connection", http.StatusInternalServerError)
		lobby.lock.Unlock()
		return
	}

	receiver, err := lobby.game.AddPlayer(playerID)
	if err != nil {
		log.Println(err)
		http.Error(res, "unable to join game", http.StatusConflict)
		lobby.lock.Unlock()
		return
	}

	player = &Player{
		id:     playerID,
		socket: socket,
		active: true,
		lock:   new(sync.RWMutex),
	}
	lobby.players[playerID] = player
	go player.listen(receiver)

	res.Write([]byte("joined lobby"))

	lobby.lock.Unlock()
	lobby.wg.Done()
}

type lobbyCreationHandler struct {
	games map[string]GameConstructor
}

// Returns a handler for creating lobbies (for servers with public lobby creation).
func (handler lobbyCreationHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	params, ok := checkParams(req, "id", "game")
	if !ok {
		http.Error(res, "insufficient query parameters", http.StatusBadRequest)
		return
	}

	id, err := url.QueryUnescape(params.Get("id"))
	if err != nil {
		http.Error(res, "invalid lobby ID provided", http.StatusBadRequest)
		return
	}

	gameTitle, err := url.QueryUnescape(params.Get("game"))
	if err != nil {
		http.Error(res, "invalid game title provided", http.StatusBadRequest)
	}

	gameConstructor, ok := handler.games[gameTitle]
	if !ok {
		http.Error(res, "invalid game title provided", http.StatusBadRequest)
		return
	}

	_, err = New(id, gameConstructor)
	if err != nil {
		http.Error(res, "error creating lobby", http.StatusInternalServerError)
		return
	}

	res.Write([]byte("lobby created"))
}

type gameListHandler struct {
	gameTitles []string
}

// Returns a handler for showing the list of games supported by the server.
func (handler gameListHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	jsonResponse, err := json.Marshal(handler.gameTitles)
	if err != nil {
		http.Error(res, "error fetching list of games", http.StatusInternalServerError)
		return
	}

	res.Write(jsonResponse)
}
