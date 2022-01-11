package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/immerse-ntnu/hermannia/server/app"
	"github.com/immerse-ntnu/hermannia/server/lobby"
)

// Launches a game server with a public endpoint for creating lobbies.
func main() {
	fmt.Println("Server started...")

	lobby.RegisterEndpoints(nil)
	lobby.RegisterLobbyCreationEndpoints(nil, app.Games)

	port := "7000"
	fmt.Printf("Listening on port %s...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Fatal(err)
}
