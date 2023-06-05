package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"

	"hermannm.dev/bfh-server/api"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/bfh-server/lobby"
	"hermannm.dev/ipfinder"
)

var availableGames = []api.GameOption{
	{
		DescriptiveName:     "The Battle for Hermannia (5 players)",
		BoardConfigFileName: "bfh_5players",
	},
}

const defaultPort string = "8000"

func main() {
	local, port := getCommandLineFlags()

	lobbyRegistry := lobby.NewLobbyRegistry()
	api.RegisterEndpoints(http.DefaultServeMux, lobbyRegistry)

	if local {
		selectedGame := selectGame(availableGames)
		createLobby(selectedGame, lobbyRegistry)
		printIPs(port)
	} else {
		api.RegisterLobbyCreationEndpoints(http.DefaultServeMux, lobbyRegistry, availableGames)
	}

	fmt.Printf("Listening on port %s...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Fatal(err)
}

func getCommandLineFlags() (local bool, port string) {
	flag.BoolVar(
		&local, "local", false,
		"Disable public endpoints for creating new lobbies",
	)
	flag.StringVar(
		&port, "port", defaultPort,
		"The port on which the server should handle requests",
	)
	flag.Parse()
	return local, port
}

func selectGame(availableGames []api.GameOption) string {
	fmt.Println("Available games:")

	for index, game := range availableGames {
		fmt.Printf("[%d] %s\n", index, game.DescriptiveName)
	}
	fmt.Println()

	var selectedGame string
	for {
		fmt.Print("Select game (type number from above list): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		index, err := strconv.Atoi(input)
		if err != nil || index < 0 || index >= len(availableGames) {
			fmt.Println("Invalid game selection, try again!")
			continue
		}

		selection := availableGames[index]
		selectedGame = selection.BoardConfigFileName
		fmt.Printf("Selected %s!\n\n", selection.DescriptiveName)
		break
	}

	return selectedGame
}

func createLobby(selectedGame string, lobbyRegistry *lobby.LobbyRegistry) {
	var lobbyName string
	for {
		fmt.Print("Type name of lobby: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		lobbyName = scanner.Text()

		lobby, err := lobby.New(lobbyName, selectedGame, game.DefaultOptions())
		if err != nil {
			fmt.Printf("Got error: '%s', try again!\n", err.Error())
			continue
		}

		if err := lobbyRegistry.RegisterLobby(lobby); err != nil {
			fmt.Printf("Got error: '%s', try again!\n", err.Error())
			continue
		}

		break
	}

	fmt.Printf("Lobby '%s' created!\n\n", lobbyName)
}

func printIPs(port string) {
	publicIP, publicErr := ipfinder.FindPublicIP()
	localIPs, localErr := ipfinder.FindLocalIPs()

	fmt.Println("Game clients should now see lobby at:")

	writer := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	if publicErr == nil {
		fmt.Fprintf(writer, "%s:%s\t(if port forwarding)\n", publicIP, port)
	} else {
		fmt.Printf("[Error finding public IP] %s\n", publicErr.Error())
	}

	if localErr == nil {
		for _, ips := range localIPs {
			for _, localIP := range ips {
				fmt.Fprintf(writer, "%s:%s\t(if on the same network)\n", localIP, port)
			}
		}
	} else {
		fmt.Printf("[Error finding local IPs] %s\n", publicErr.Error())
	}

	writer.Flush()
	fmt.Println()
}
