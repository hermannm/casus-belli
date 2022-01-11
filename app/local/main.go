package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/immerse-ntnu/hermannia/server/app"
	"github.com/immerse-ntnu/hermannia/server/lobby"
)

// Launches a game server that runs a single lobby and game.
// Starts by configuring the lobby and game through a CLI, then listens.
func main() {
	fmt.Println("Server started...")
	fmt.Println()

	game := selectGame(app.Games)
	fmt.Println()

	createLobby(game)
	fmt.Println()

	lobby.RegisterEndpoints(nil)

	port := "7000"
	fmt.Printf("Listening on port %s...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Fatal(err)
}

// CLI utility for selecting a game from a list of supported games.
// Does not return until a valid game selection is made.
func selectGame(games map[string]lobby.GameConstructor) lobby.GameConstructor {
	fmt.Print("Available games:\n")

	gameTitles := make([]string, 0)
	for key := range games {
		gameTitles = append(gameTitles, key)
	}
	for index, title := range gameTitles {
		fmt.Printf("%d: %s\n", index, title)
	}
	fmt.Println()

	var selectedTitle string
	for {
		fmt.Print("Select game (type index or title): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		selectedTitle = scanner.Text()

		index, err := strconv.Atoi(selectedTitle)
		if err == nil {
			selectedTitle = gameTitles[index]
			break
		}

		_, ok := games[selectedTitle]
		if ok {
			break
		}

		fmt.Println("Invalid game selection, try again!")
	}

	fmt.Printf("Selected \"%s\"!\n", selectedTitle)
	return games[selectedTitle]
}

// CLI utility for creating a lobby.
// Does not return until a valid lobby is created.
func createLobby(game lobby.GameConstructor) {
	var lobbyName string
	for {
		fmt.Print("Type name of lobby: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		lobbyName = scanner.Text()

		_, err := lobby.New(lobbyName, game)
		if err == nil {
			break
		}

		fmt.Printf("Got error: \"%s\", try again!\n", err.Error())
	}

	fmt.Printf("Lobby \"%s\" created!\n", lobbyName)
}
