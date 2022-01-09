package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/immerse-ntnu/hermannia/server/interfaces"
)

// Registers handlers for the lobby API endpoints.
func Start(address string) {
	// Endpoint for clients to join a given lobby.
	// Takes query parameters "lobby" (name of the lobby) and "player" (the player ID that the client wants to claim).
	http.HandleFunc("/join", addPlayer)

	// Endpoint for clients to view info about a single lobby.
	// Takes query parameter "lobby" (name of the lobby).
	http.HandleFunc("/info", getLobby)

	// Endpoint for clients to view info about all lobbies on the server.
	http.HandleFunc("/all", getLobbies)

	http.ListenAndServe(address, nil)
}

func StartPublic(address string, games map[string]interfaces.GameConstructor) {
	// Endpoint for clients to create their own lobbies if the server is set to enable that.
	// Takes query parameters "id" (unique name of the lobby) and "playerIDs".
	http.HandleFunc("/new", createLobbyHandler(games))

	Start(address)
}

// Checks the given request for the existence of the provided parameter keys.
// If all exist, returns the parameters, otherwise returns ok = false.
func checkParams(res http.ResponseWriter, req *http.Request, keys ...string) (
	params url.Values, ok bool,
) {
	params = req.URL.Query()

	for _, key := range keys {
		if !params.Has(key) {
			http.Error(res, "insufficient query parameters", http.StatusBadRequest)
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
	params, ok := checkParams(res, req, "lobby")
	if !ok {
		return
	}

	lobbyID := params.Get("lobby")
	lobby, ok := lobbies[lobbyID]
	if !ok {
		http.Error(res, "no lobby with id \""+lobbyID+"\"", http.StatusBadRequest)
		return
	}

	info, err := json.Marshal(lobbyInfo{
		ID:                 lobby.ID,
		AvailablePlayerIDs: lobby.AvailablePlayerIDs(),
	})
	if err != nil {
		http.Error(res, "error in reading lobby \""+lobbyID+"\"", http.StatusInternalServerError)
	}

	res.Write(info)
}

// Handler for returning information about all available lobbies.
func getLobbies(res http.ResponseWriter, req *http.Request) {
	lobbyInfoList := make([]lobbyInfo, 0, len(lobbies))

	for _, lobby := range lobbies {
		lobbyInfoList = append(lobbyInfoList, lobbyInfo{
			ID:                 lobby.ID,
			AvailablePlayerIDs: lobby.AvailablePlayerIDs(),
		})
	}

	info, err := json.Marshal(lobbyInfoList)
	if err != nil {
		http.Error(res, "error in reading lobby fetching lobby list", http.StatusInternalServerError)
	}

	res.Write(info)
}

// Handler for adding a player to a lobby.
func addPlayer(res http.ResponseWriter, req *http.Request) {
	params, ok := checkParams(res, req, "lobby", "player")
	if !ok {
		return
	}

	lobbyID := params.Get("lobby")
	lobby, ok := lobbies[lobbyID]
	if !ok {
		http.Error(res, "no lobby with ID "+lobbyID+" exists", http.StatusBadRequest)
	}

	playerID := params.Get("player")
	lobby.Mut.Lock()
	conn, ok := lobby.Connections[playerID]
	if !ok {
		http.Error(res, "invalid player ID", http.StatusBadRequest)
		lobby.Mut.Unlock()
		return
	}
	if conn.isActive() {
		http.Error(res, "player ID already taken", http.StatusConflict)
		lobby.Mut.Unlock()
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
		lobby.Mut.Unlock()
		return
	}

	receiver, err := lobby.Game.AddPlayer(playerID)
	if err != nil {
		log.Println(err)
		http.Error(res, "unable to join game", http.StatusConflict)
		lobby.Mut.Unlock()
		return
	}

	conn = &Connection{
		Socket:   socket,
		Active:   true,
		Receiver: receiver,
	}
	lobby.Connections[playerID] = conn
	go conn.Listen()

	res.Write([]byte("joined lobby"))

	lobby.Mut.Unlock()
	lobby.WG.Done()
}

// Returns a handler for creating lobbies (for servers with public lobby creation).
func createLobbyHandler(games map[string]interfaces.GameConstructor) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		params, ok := checkParams(res, req, "id", "playerIDs", "game")
		if !ok {
			return
		}

		id := params.Get("id")
		if _, ok := lobbies[id]; ok {
			http.Error(res, "lobby with ID \""+id+"\" already exists", http.StatusConflict)
			return
		}

		playerIDs := params["playerIDs"]
		if len(playerIDs) < 2 {
			http.Error(res, "at least 2 player IDs must be provided to lobby", http.StatusBadRequest)
			return
		}

		lobby := NewLobby(id, playerIDs)

		gameConstructor, ok := games[params.Get("game")]
		if !ok {
			http.Error(res, "invalid game descriptor provided", http.StatusBadRequest)
		}

		game, err := gameConstructor(playerIDs, &lobby, nil)
		if err != nil {
			http.Error(res, "error creating game", http.StatusInternalServerError)
		}

		lobby.Game = game

		err = RegisterLobby(&lobby)
		if err != nil {
			http.Error(res, "error creating lobby", http.StatusInternalServerError)
			return
		}

		res.Write([]byte("lobby created"))
	}
}
