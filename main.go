package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"hermannm.dev/bfh-server/api"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/bfh-server/game/boardconfig"
	"hermannm.dev/bfh-server/lobby"
	"hermannm.dev/ipfinder"
	"hermannm.dev/wrap"
)

const defaultPort string = "8000"

func main() {
	local, port := getCommandLineFlags()

	availableBoards, err := boardconfig.GetAvailableBoards()
	if err != nil {
		fmt.Println(wrap.Error(err, "failed to get available boards for game server"))
		os.Exit(1)
	}

	lobbyRegistry := lobby.NewLobbyRegistry()
	lobbyAPI := api.NewLobbyAPI(http.DefaultServeMux, lobbyRegistry, availableBoards)

	if local {
		selectedBoardID := selectBoard(availableBoards)
		createLobby(selectedBoardID, lobbyRegistry)
		printIPs(port)
	} else {
		lobbyAPI.RegisterLobbyCreationEndpoints()
	}

	fmt.Printf("Listening on port %s...", port)
	if err := lobbyAPI.ListenAndServe(fmt.Sprintf(":%s", port)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getCommandLineFlags() (local bool, port string) {
	flag.BoolVar(&local, "local", false, "Disable public endpoints for creating new lobbies")
	flag.StringVar(
		&port,
		"port",
		defaultPort,
		"The port on which the server should handle requests",
	)
	flag.Parse()
	return local, port
}

func selectBoard(availableBoards []boardconfig.BoardInfo) string {
	fmt.Println("Available boards:")

	for index, board := range availableBoards {
		fmt.Printf("[%d] %s\n", index, board.DescriptiveName)
	}
	fmt.Println()

	var selectedBoardID string
	for {
		fmt.Print("Select board (type number from above list): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		index, err := strconv.Atoi(input)
		if err != nil || index < 0 || index >= len(availableBoards) {
			fmt.Println("Invalid board selection, try again!")
			continue
		}

		selection := availableBoards[index]
		selectedBoardID = selection.ID
		fmt.Printf("Selected %s!\n\n", selection.DescriptiveName)
		break
	}

	return selectedBoardID
}

func createLobby(selectedBoardID string, lobbyRegistry *lobby.LobbyRegistry) {
	var lobbyName string
	for {
		fmt.Print("Type name of lobby: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		lobbyName = scanner.Text()

		lobby, err := lobby.New(lobbyName, selectedBoardID, game.DefaultOptions())
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
	fmt.Println("Game clients should now see lobby at:")

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	publicIP, publicErr := ipfinder.FindPublicIP(ctx)
	localIPs, localErr := ipfinder.FindLocalIPs()

	writer := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	if publicErr == nil {
		fmt.Fprintf(writer, "%s:%s\t(if port forwarding)\n", publicIP, port)
	} else {
		fmt.Printf("[Error finding public IP] %s\n", publicErr.Error())
	}

	if localErr == nil {
		for _, ip := range localIPs {
			fmt.Fprintf(writer, "%s:%s\t(if on the same network)\n", ip.Address.String(), port)
		}
	} else {
		fmt.Printf("[Error finding local IPs] %s\n", localErr.Error())
	}

	writer.Flush()
	fmt.Println()
}
