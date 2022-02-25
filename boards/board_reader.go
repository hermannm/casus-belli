package boards

import (
	"embed"
	"encoding/json"
	"fmt"

	"hermannm.dev/bfh-server/game"
)

// boards embeds the json files of boards from this folder.
//go:embed hermannia_5players.json
var boards embed.FS

// Utility type for json unmarshaling.
type board struct {
	Nations   map[string][]landArea `json:"nations"`
	Seas      []sea                 `json:"seas"`
	Neighbors []neighbor            `json:"neighbors"`
}

// Utility type for json unmarshaling.
type landArea struct {
	Name   string `json:"name"`
	Forest bool   `json:"forest"`
	Castle bool   `json:"castle"`
	Home   string `json:"home"`
}

// Utility type for json unmarshaling.
type sea struct {
	Name string `json:"name"`
}

// Utility type for json unmarshaling.
type neighbor struct {
	Area1      string `json:"area1"`
	Area2      string `json:"area2"`
	River      bool   `json:"river"`
	Cliffs     bool   `json:"cliffs"`
	DangerZone string `json:"dangerZone"`
}

// Reads and constructs the board matching the given map name and number of players.
func ReadBoard(boardName string) (game.Board, error) {
	content, err := boards.ReadFile(fmt.Sprintf("%s.json", boardName))
	if err != nil {
		return nil, err
	}

	var jsonBoard board

	err = json.Unmarshal(content, &jsonBoard)
	if err != nil {
		return nil, err
	}

	board := make(game.Board)

	for nation, areas := range jsonBoard.Nations {
		for _, landArea := range areas {
			area := game.Area{
				Name:             landArea.Name,
				Nation:           nation,
				Control:          game.Player(landArea.Home),
				Home:             game.Player(landArea.Home),
				Forest:           landArea.Forest,
				Castle:           landArea.Castle,
				Neighbors:        make([]game.Neighbor, 0),
				IncomingMoves:    make([]game.Order, 0),
				IncomingSupports: make([]game.Order, 0),
			}

			board[area.Name] = area
		}
	}

	for _, sea := range jsonBoard.Seas {
		area := game.Area{
			Name:             sea.Name,
			Neighbors:        make([]game.Neighbor, 0),
			IncomingMoves:    make([]game.Order, 0),
			IncomingSupports: make([]game.Order, 0),
		}

		board[area.Name] = area
	}

	for _, neighbor := range jsonBoard.Neighbors {
		area1, ok1 := board[neighbor.Area1]
		area2, ok2 := board[neighbor.Area2]

		if !ok1 || !ok2 {
			return nil, fmt.Errorf(
				"error in board config: neighbor relation %s <-> %s",
				neighbor.Area1,
				neighbor.Area2,
			)
		}

		area1.Neighbors = append(area1.Neighbors, game.Neighbor{
			Name:        area2.Name,
			AcrossWater: neighbor.River || (area1.Sea && !area2.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})

		area2.Neighbors = append(area2.Neighbors, game.Neighbor{
			Name:        area1.Name,
			AcrossWater: neighbor.River || (area2.Sea && !area1.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})
	}

	return board, nil
}
