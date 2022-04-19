package main

import (
	"fmt"
	"log"
	"net/http"

	server "hermannm.dev/bfh-server"
	"hermannm.dev/bfh-server/lobby"
)

// Launches a game server with a public endpoint for creating lobbies.
func main() {
	fmt.Println("Server started...")

	lobby.RegisterEndpoints(nil)
	lobby.RegisterLobbyCreationEndpoints(nil, server.Games)

	port := "8000"
	fmt.Printf("Listening on port %s...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Fatal(err)
}
