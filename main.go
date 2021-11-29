package main

import (
	"fmt"
	"immerse-ntnu/hermannia/server/game"
	"immerse-ntnu/hermannia/server/game/setup"
)

func printBoard(board game.Board, areas map[string]*game.Unit, neighbors bool) {
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

		if area.Unit != nil {
			fmt.Println("Unit:", area.Unit.Color, area.Unit.Type)
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

		if len(area.Combats) > 0 {
			fmt.Println("\nCOMBATS:")
			for _, combat := range area.Combats {
				combatString := ""

				for _, result := range combat {
					resultString := fmt.Sprintf("%s: %d ( ", result.Player, result.Total)

					for _, mod := range result.Parts {
						resultString += fmt.Sprintf("%d %s ", mod.Value, mod.Type)
						if mod.SupportFrom != "" {
							resultString += string(mod.SupportFrom) + " "
						}
					}

					resultString += ") "

					combatString += resultString
				}

				fmt.Println(combatString)
			}
		}

		fmt.Print("\n-----------------------------------\n\n")
	}
}

func main() {
	board, err := setup.ReadBoard(5)

	if err != nil {
		fmt.Println(err.Error())
	}

	testTransportCombat(board)
}

func testTransportCombat(board game.Board) {
	areas := map[string]*game.Unit{
		"Worp": {
			Type:  game.Footman,
			Color: "green",
		},
		"Mare Gond": {
			Type:  game.Ship,
			Color: "green",
		},
		"Mare Elle": {
			Type:  game.Ship,
			Color: "red",
		},
		"Zona": nil,
	}

	for key, unit := range areas {
		if unit != nil {
			board[key].Unit = unit
			if !board[key].Sea {
				board[key].Control = unit.Color
			}
		}
	}

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

	fmt.Print("---BEFORE---\n\n")
	printBoard(board, areas, false)

	board.Resolve(&round)

	fmt.Print("---AFTER---\n\n")
	printBoard(board, areas, false)
}
