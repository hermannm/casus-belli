package boards

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/hermannm/bfh-server/game"
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
		for _, jsonArea := range areas {
			area := game.Area{
				Name:             jsonArea.Name,
				Nation:           nation,
				Control:          game.Player(jsonArea.Home),
				Home:             game.Player(jsonArea.Home),
				Forest:           jsonArea.Forest,
				Castle:           jsonArea.Castle,
				Neighbors:        make([]game.Neighbor, 0),
				IncomingMoves:    make([]*game.Order, 0),
				IncomingSupports: make([]*game.Order, 0),
				Battles:          make([]game.Battle, 0),
			}

			board[area.Name] = &area
		}
	}

	for _, sea := range jsonBoard.Seas {
		area := game.Area{
			Name:             sea.Name,
			Neighbors:        make([]game.Neighbor, 0),
			IncomingMoves:    make([]*game.Order, 0),
			IncomingSupports: make([]*game.Order, 0),
			Battles:          make([]game.Battle, 0),
		}

		board[area.Name] = &area
	}

	for _, jsonNeighbor := range jsonBoard.Neighbors {
		area1, ok1 := board[jsonNeighbor.Area1]
		area2, ok2 := board[jsonNeighbor.Area2]

		if !ok1 || !ok2 {
			return nil, fmt.Errorf(
				"error in board config: neighbor relation %s <-> %s",
				jsonNeighbor.Area1,
				jsonNeighbor.Area2,
			)
		}

		area1.Neighbors = append(area1.Neighbors, game.Neighbor{
			Area:       area2,
			River:      jsonNeighbor.River,
			Cliffs:     jsonNeighbor.Cliffs,
			DangerZone: jsonNeighbor.DangerZone,
		})

		area2.Neighbors = append(area2.Neighbors, game.Neighbor{
			Area:       area1,
			River:      jsonNeighbor.River,
			Cliffs:     jsonNeighbor.Cliffs,
			DangerZone: jsonNeighbor.DangerZone,
		})
	}

	return board, nil
}
