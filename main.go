package main

import (
	"fmt"

	"github.com/immerse-ntnu/hermannia/server/game"
	"github.com/immerse-ntnu/hermannia/server/game/setup"
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
				neighborString := neighbor.Area.Name
				if neighbor.River {
					neighborString += " (river)"
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

		if len(area.Battles) > 0 {
			fmt.Println("\nBATTLES:")
			for _, battle := range area.Battles {
				battleString := ""

				for _, result := range battle {
					resultString := fmt.Sprintf("%s: %d ( ", result.Player, result.Total)

					for _, mod := range result.Parts {
						resultString += fmt.Sprintf("%d %s ", mod.Value, mod.Type)
						if mod.SupportFrom != "" {
							resultString += string(mod.SupportFrom) + " "
						}
					}

					resultString += ") "

					battleString += resultString
				}

				fmt.Println(battleString)
			}
		}

		fmt.Print("\n-----------------------------------\n\n")
	}
}

func adjustBoard(board game.Board, areas map[string]game.Unit) {
	for key, unit := range areas {
		if unit.Type != game.NoUnit {
			board[key].Unit = unit
			if !board[key].Sea {
				board[key].Control = unit.Player
			}
		}
	}
}

func printResolvePrint(board game.Board, areas map[string]game.Unit, round *game.Round) {
	fmt.Print("---BEFORE---\n\n")
	printBoard(board, areas, false)

	board.Resolve(round)

	fmt.Print("---AFTER---\n\n")
	printBoard(board, areas, false)
}

func main() {
	board, err := setup.ReadBoard(5)

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
		FirstOrders: []*game.Order{
			{
				Type:   game.Move,
				Player: "green",
				From:   board["Winde"],
				To:     board["Fond"],
			},
			{
				Type:   game.Transport,
				Player: "green",
				From:   board["Mare Gond"],
			},
			{
				Type:   game.Transport,
				Player: "green",
				From:   board["Mare Elle"],
			},
			{
				Type:   game.Transport,
				Player: "green",
				From:   board["Mare Ovond"],
			},
			{
				Type:   game.Move,
				Player: "red",
				From:   board["Mare Duna"],
				To:     board["Mare Gond"],
			},
			{
				Type:   game.Move,
				Player: "red",
				From:   board["Mare Furie"],
				To:     board["Mare Elle"],
			},
		},
	}

	printResolvePrint(board, areas, &round)
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
		FirstOrders: []*game.Order{
			{
				Type:   game.Move,
				Player: "green",
				From:   board["Worp"],
				To:     board["Zona"],
			},
			{
				Type:   game.Transport,
				Player: "green",
				From:   board["Mare Gond"],
			},
			{
				Type:   game.Move,
				Player: "red",
				From:   board["Mare Elle"],
				To:     board["Mare Gond"],
			},
		},
	}

	printResolvePrint(board, areas, &round)
}
