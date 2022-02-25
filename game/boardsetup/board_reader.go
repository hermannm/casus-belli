package boardsetup

import (
	"embed"
	"encoding/json"
	"fmt"

	"hermannm.dev/bfh-server/game/board"
)

// boards embeds the json files of boards from this folder.
//go:embed hermannia_5players.json
var boards embed.FS

// Utility type for json unmarshaling.
type jsonBoard struct {
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
func ReadBoard(boardName string) (board.Board, error) {
	content, err := boards.ReadFile(fmt.Sprintf("%s.json", boardName))
	if err != nil {
		return nil, err
	}

	var jsonBrd jsonBoard

	err = json.Unmarshal(content, &jsonBrd)
	if err != nil {
		return nil, err
	}

	brd := make(board.Board)

	for nation, areas := range jsonBrd.Nations {
		for _, landArea := range areas {
			area := board.Area{
				Name:             landArea.Name,
				Nation:           nation,
				Control:          board.Player(landArea.Home),
				Home:             board.Player(landArea.Home),
				Forest:           landArea.Forest,
				Castle:           landArea.Castle,
				Neighbors:        make([]board.Neighbor, 0),
				IncomingMoves:    make([]board.Order, 0),
				IncomingSupports: make([]board.Order, 0),
			}

			brd[area.Name] = area
		}
	}

	for _, sea := range jsonBrd.Seas {
		area := board.Area{
			Name:             sea.Name,
			Neighbors:        make([]board.Neighbor, 0),
			IncomingMoves:    make([]board.Order, 0),
			IncomingSupports: make([]board.Order, 0),
		}

		brd[area.Name] = area
	}

	for _, neighbor := range jsonBrd.Neighbors {
		area1, ok1 := brd[neighbor.Area1]
		area2, ok2 := brd[neighbor.Area2]

		if !ok1 || !ok2 {
			return nil, fmt.Errorf(
				"error in board config: neighbor relation %s <-> %s",
				neighbor.Area1,
				neighbor.Area2,
			)
		}

		area1.Neighbors = append(area1.Neighbors, board.Neighbor{
			Name:        area2.Name,
			AcrossWater: neighbor.River || (area1.Sea && !area2.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})

		area2.Neighbors = append(area2.Neighbors, board.Neighbor{
			Name:        area1.Name,
			AcrossWater: neighbor.River || (area2.Sea && !area1.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})
	}

	return brd, nil
}
