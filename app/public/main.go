package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hermannm/bfh-server/app"
	"github.com/hermannm/bfh-server/lobby"
)

// Launches a game server with a public endpoint for creating lobbies.
func main() {
	fmt.Println("Server started...")

	lobby.RegisterEndpoints(nil)
	lobby.RegisterLobbyCreationEndpoints(nil, app.Games)

	port := "7000"
	fmt.Printf("Listening on port %s...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Fatal(err)
}
