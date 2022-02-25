package main

import (
	"fmt"

	"hermannm.dev/bfh-server/boards"
	"hermannm.dev/bfh-server/game"
)

func printBoard(board game.Board, areas map[string]game.Unit, neighbors bool) {
	for _, area := range board {
		if _, ok := areas[area.Name]; !ok {
			continue
		}

		areaString := area.Name
		if area.Control != "" {
			areaString += fmt.Sprintf(" (control: %s)", area.Control)
		}
		if area.Home != "" {
			areaString += fmt.Sprintf(" (home: %s)", area.Home)
		}
		if area.Sea {
			areaString += " (sea) "
		}
		if area.Forest {
			areaString += " (forest)"
		}
		if area.Castle {
			areaString += " (castle)"
		}
		fmt.Println(areaString)

		if !area.IsEmpty() {
			fmt.Println("Unit:", area.Unit.Player, area.Unit.Type)
		}

		if neighbors {
			fmt.Println("\nNEIGHBORS:")
			for _, neighbor := range area.Neighbors {
				neighborString := neighbor.Name
				if neighbor.AcrossWater {
					neighborString += " (across water)"
				}
				if neighbor.Cliffs {
					neighborString += " (cliffs)"
				}
				if neighbor.DangerZone != "" {
					neighborString += fmt.Sprintf(" (danger: %s)", neighbor.DangerZone)
				}
				fmt.Println(neighborString)
			}
		}

		fmt.Print("\n-----------------------------------\n\n")
	}
}

func adjustBoard(board game.Board, areas map[string]game.Unit) {
	for areaName, unit := range areas {
		if unit.Type != "" {
			area := board[areaName]
			area.Unit = unit
			if !area.Sea {
				area.Control = unit.Player
			}
			board[areaName] = area
		}
	}
}

func printResolvePrint(board game.Board, areas map[string]game.Unit, round game.Round) {
	fmt.Print("---BEFORE---\n\n")
	printBoard(board, areas, false)

	board.Resolve(round)

	fmt.Print("---AFTER---\n\n")
	printBoard(board, areas, false)
}

func main() {
	board, err := boards.ReadBoard("hermannia_5players")

	if err != nil {
		fmt.Println(err.Error())
	}

	testTransportWithDangerZone(board)
}

func testTransportWithDangerZone(board game.Board) {
	areas := map[string]game.Unit{
		"Winde": {
			Type:   game.Footman,
			Player: "green",
		},
		"Mare Gond": {
			Type:   game.Ship,
			Player: "green",
		},
		"Mare Elle": {
			Type:   game.Ship,
			Player: "green",
		},
		"Mare Ovond": {
			Type:   game.Ship,
			Player: "green",
		},
		"Mare Duna": {
			Type:   game.Ship,
			Player: "red",
		},
		"Mare Furie": {
			Type:   game.Ship,
			Player: "red",
		},
		"Fond": {},
	}

	adjustBoard(board, areas)

	round := game.Round{
		Season: game.Spring,
		FirstOrders: []game.Order{
			{
				Type:   game.Move,
				Player: "green",
				From:   "Winde",
				To:     "Fond",
				Unit:   board["Winde"].Unit,
			},
			{
				Type:   game.Transport,
				Player: "green",
				From:   "Mare Gond",
				Unit:   board["Mare Gond"].Unit,
			},
			{
				Type:   game.Transport,
				Player: "green",
				From:   "Mare Elle",
				Unit:   board["Mare Elle"].Unit,
			},
			{
				Type:   game.Transport,
				Player: "green",
				From:   "Mare Ovond",
				Unit:   board["Mare Ovond"].Unit,
			},
			{
				Type:   game.Move,
				Player: "red",
				From:   "Mare Duna",
				To:     "Mare Gond",
				Unit:   board["Mare Gond"].Unit,
			},
			{
				Type:   game.Move,
				Player: "red",
				From:   "Mare Furie",
				To:     "Mare Elle",
				Unit:   board["Mare Furie"].Unit,
			},
		},
	}

	printResolvePrint(board, areas, round)
}

func testTransportBattle(board game.Board) {
	areas := map[string]game.Unit{
		"Worp": {
			Type:   game.Footman,
			Player: "green",
		},
		"Mare Gond": {
			Type:   game.Ship,
			Player: "green",
		},
		"Mare Elle": {
			Type:   game.Ship,
			Player: "red",
		},
		"Zona": {},
	}

	adjustBoard(board, areas)

	round := game.Round{
		Season: game.Spring,
		FirstOrders: []game.Order{
			{
				Type:   game.Move,
				Player: "green",
				From:   "Worp",
				To:     "Zona",
				Unit:   board["Worp"].Unit,
			},
			{
				Type:   game.Transport,
				Player: "green",
				From:   "Mare Gond",
				Unit:   board["Mare Gond"].Unit,
			},
			{
				Type:   game.Move,
				Player: "red",
				From:   "Mare Elle",
				To:     "Mare Gond",
				Unit:   board["Mare Gond"].Unit,
			},
		},
	}

	printResolvePrint(board, areas, round)
}
