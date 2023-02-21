package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"

	server "hermannm.dev/bfh-server"
	"hermannm.dev/bfh-server/lobby"
	"hermannm.dev/ipfinder"
)

// Launches a game server that runs a single lobby and game.
// Starts by configuring the lobby and game through a CLI, then listens.
func main() {
	fmt.Println("Server started...")
	fmt.Println()

	game := selectGame(server.Games)
	fmt.Println()

	createLobby(game)
	fmt.Println()

	lobby.RegisterEndpoints(nil)

	port := "8000"

	printIPs(port)

	fmt.Println("Listening...")
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Fatal(err)
}

// Prompts the user to select a game from a list of supported games.
// Does not return until a valid game selection is made.
func selectGame(games map[string]lobby.GameConstructor) lobby.GameConstructor {
	fmt.Print("Available games:\n")

	var gameTitles []string
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

// Prompts the user to choose a name for the lobby, then creates it.
// Constructs the lobby's game instance with the given constructor.
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

// Prints out IPs that may be used to connect to the lobby.
func printIPs(port string) {
	publicIP, publicErr := ipfinder.FindPublicIP()
	localIPs, localErr := ipfinder.FindLocalIPs()

	fmt.Println("Lobby can now be joined at:")

	writer := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	if publicErr == nil {
		fmt.Fprintf(writer, "%s:%s/join\t(if port forwarding)\n", publicIP, port)
	} else {
		fmt.Printf("[Error finding public IP] %s\n", publicErr.Error())
	}
	if localErr == nil {
		for _, ips := range localIPs {
			for _, localIP := range ips {
				fmt.Fprintf(writer, "%s:%s/join\t(if on the same network)\n", localIP, port)
			}
		}
	} else {
		fmt.Printf("[Error finding local IPs] %s\n", publicErr.Error())
	}
	writer.Flush()

	fmt.Println()
}
