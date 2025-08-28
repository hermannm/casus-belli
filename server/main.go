package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
	"hermannm.dev/ipfinder"

	"hermannm.dev/casus-belli/server/api"
	"hermannm.dev/casus-belli/server/game"
	"hermannm.dev/casus-belli/server/lobby"
)

const defaultPort string = "8000"

func main() {
	log.SetDefault(devlog.NewHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug}))

	ctx := context.Background()

	local, devMode, port := getCommandLineFlags()

	availableBoards, err := game.GetAvailableBoards()
	if err != nil {
		log.Error(ctx, err, "Failed to get available boards for game server")
		os.Exit(1)
	}

	lobbyRegistry := lobby.NewLobbyRegistry()
	lobbyAPI := api.NewLobbyAPI(http.DefaultServeMux, lobbyRegistry, availableBoards)

	if local || devMode {
		selectedBoard := selectBoard(availableBoards)
		createLobby(selectedBoard, lobbyRegistry, devMode)
		printIPs(ctx, port)
	} else {
		lobbyAPI.RegisterLobbyCreationEndpoints()
	}

	log.Infof(ctx, "Listening on port %s...", port)
	if err := lobbyAPI.ListenAndServe(fmt.Sprintf(":%s", port)); err != nil {
		log.Error(ctx, err, "Server stopped")
		os.Exit(1)
	}
}

func getCommandLineFlags() (local bool, devMode bool, port string) {
	flag.BoolVar(&local, "local", false, "Disable public endpoints for creating new lobbies")
	flag.BoolVar(
		&devMode,
		"dev",
		false,
		"Allows for creating single-player lobbies for development",
	)
	flag.StringVar(
		&port,
		"port",
		defaultPort,
		"The port on which the server should handle requests",
	)
	flag.Parse()
	return local, devMode, port
}

//nolint:forbidigo
func selectBoard(availableBoards []game.BoardInfo) game.BoardInfo {
	if len(availableBoards) == 1 {
		return availableBoards[0]
	}

	fmt.Println("Available boards:")

	for index, board := range availableBoards {
		fmt.Printf("[%d] %s\n", index, board.Name)
	}
	fmt.Println()

	var selectedBoard game.BoardInfo
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

		selectedBoard = availableBoards[index]
		fmt.Printf("Selected %s!\n\n", selectedBoard.Name)
		break
	}

	return selectedBoard
}

//nolint:forbidigo
func createLobby(
	selectedBoard game.BoardInfo,
	lobbyRegistry *lobby.LobbyRegistry,
	devMode bool,
) {
	var lobbyName string
	for {
		fmt.Print("Lobby name: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		lobbyName = scanner.Text()

		var customFactions []game.PlayerFaction
		if devMode {
			fmt.Println()
			fmt.Println("[dev] Playable factions:")
			for i, faction := range selectedBoard.PlayerFactions {
				fmt.Printf("  [%d] %s\n", i, faction)
			}

			fmt.Print("Select faction (type number from above list): ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			index, err := strconv.Atoi(scanner.Text())
			if err != nil || index < 0 || index >= len(selectedBoard.PlayerFactions) {
				fmt.Println("Invalid faction selection, try again!")
				continue
			}
			fmt.Println()

			customFactions = []game.PlayerFaction{selectedBoard.PlayerFactions[index]}
		}

		if err := lobbyRegistry.CreateLobby(
			lobbyName,
			selectedBoard.ID,
			true,
			customFactions,
		); err != nil {
			fmt.Printf("Got error: '%s', try again!\n", err.Error())
			continue
		}

		break
	}

	fmt.Printf("Lobby '%s' created!\n\n", lobbyName)
}

//nolint:forbidigo
func printIPs(ctx context.Context, port string) {
	fmt.Println("Game clients should now see lobby at:")

	writer := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	localIPs, err := ipfinder.FindLocalIPs()
	if err == nil {
		for _, ip := range localIPs {
			_, _ = fmt.Fprintf(
				writer,
				"%s:%s\t(if on the same network)\n",
				ip.Address.String(),
				port,
			)
		}
	} else {
		fmt.Printf("[Error finding local IPs] %s\n", err.Error())
	}

	ctx, cancelCtx := context.WithTimeout(ctx, 5*time.Second)
	defer cancelCtx()

	publicIP, err := ipfinder.FindPublicIP(ctx)
	if err == nil {
		_, _ = fmt.Fprintf(writer, "%s:%s\t(if port forwarding)\n", publicIP, port)
	} else {
		fmt.Printf("[Error finding public IP] %s\n", err.Error())
	}

	_ = writer.Flush()
	fmt.Println()
}
