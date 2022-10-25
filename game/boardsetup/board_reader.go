package boardsetup

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/board"
)

// boards embeds the json files of boards from this folder.
//
//go:embed bfh_5players.json
var boards embed.FS

// Utility type for json unmarshaling.
type jsonBoard struct {
	Name               string                `json:"name"`
	WinningCastleCount int                   `json:"winningCastleCount"`
	Nations            map[string][]landArea `json:"nations"`
	Seas               []seaArea             `json:"seas"`
	Neighbors          []neighbor            `json:"neighbors"`
}

// Utility type for json unmarshaling.
type landArea struct {
	Name       string `json:"name"`
	Forest     bool   `json:"forest"`
	Castle     bool   `json:"castle"`
	HomePlayer string `json:"homePlayer"`
}

// Utility type for json unmarshaling.
type seaArea struct {
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

// Reads and constructs the board matching the given ID.
func ReadBoard(boardID string) (board.Board, error) {
	content, err := boards.ReadFile(fmt.Sprintf("%s.json", boardID))
	if err != nil {
		return board.Board{}, err
	}

	var jsonBoard jsonBoard

	err = json.Unmarshal(content, &jsonBoard)
	if err != nil {
		return board.Board{}, err
	}

	if jsonBoard.WinningCastleCount <= 0 {
		return board.Board{}, errors.New("invalid winningCastleCount in board config")
	}

	brd := board.Board{
		Areas:              make(map[string]board.Area),
		Name:               jsonBoard.Name,
		WinningCastleCount: jsonBoard.WinningCastleCount,
	}

	for nation, areas := range jsonBoard.Nations {
		for _, landArea := range areas {
			area := board.Area{
				Name:              landArea.Name,
				Nation:            nation,
				ControllingPlayer: landArea.HomePlayer,
				HomePlayer:        landArea.HomePlayer,
				Forest:            landArea.Forest,
				Castle:            landArea.Castle,
				Neighbors:         make([]board.Neighbor, 0),
				IncomingMoves:     make([]board.Order, 0),
				IncomingSupports:  make([]board.Order, 0),
			}

			brd.Areas[area.Name] = area
		}
	}

	for _, sea := range jsonBoard.Seas {
		area := board.Area{
			Name:             sea.Name,
			Neighbors:        make([]board.Neighbor, 0),
			IncomingMoves:    make([]board.Order, 0),
			IncomingSupports: make([]board.Order, 0),
		}

		brd.Areas[area.Name] = area
	}

	for _, neighbor := range jsonBoard.Neighbors {
		area1, ok1 := brd.Areas[neighbor.Area1]
		area2, ok2 := brd.Areas[neighbor.Area2]

		if !ok1 || !ok2 {
			return board.Board{}, fmt.Errorf(
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
