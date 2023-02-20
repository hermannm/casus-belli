package boardsetup

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/gameboard"
)

// boards embeds the json files of boards from this folder.
//
//go:embed bfh_5players.json
var boards embed.FS

// Utility type for json unmarshaling.
type jsonBoard struct {
	Name               string                  `json:"name"`
	WinningCastleCount int                     `json:"winningCastleCount"`
	Nations            map[string][]landRegion `json:"nations"`
	Seas               []seaRegion             `json:"seas"`
	Neighbors          []neighbor              `json:"neighbors"`
}

// Utility type for json unmarshaling.
type landRegion struct {
	Name       string `json:"name"`
	Forest     bool   `json:"forest"`
	Castle     bool   `json:"castle"`
	HomePlayer string `json:"homePlayer"`
}

// Utility type for json unmarshaling.
type seaRegion struct {
	Name string `json:"name"`
}

// Utility type for json unmarshaling.
type neighbor struct {
	Region1    string `json:"region1"`
	Region2    string `json:"region2"`
	River      bool   `json:"river"`
	Cliffs     bool   `json:"cliffs"`
	DangerZone string `json:"dangerZone"`
}

// Reads and constructs the board matching the given ID.
func ReadBoard(boardID string) (gameboard.Board, error) {
	content, err := boards.ReadFile(fmt.Sprintf("%s.json", boardID))
	if err != nil {
		return gameboard.Board{}, err
	}

	var jsonBoard jsonBoard

	err = json.Unmarshal(content, &jsonBoard)
	if err != nil {
		return gameboard.Board{}, err
	}

	if jsonBoard.WinningCastleCount <= 0 {
		return gameboard.Board{}, errors.New("invalid winningCastleCount in board config")
	}

	board := gameboard.Board{
		Regions:            make(map[string]gameboard.Region),
		Name:               jsonBoard.Name,
		WinningCastleCount: jsonBoard.WinningCastleCount,
	}

	for nation, regions := range jsonBoard.Nations {
		for _, landRegion := range regions {
			region := gameboard.Region{
				Name:              landRegion.Name,
				Nation:            nation,
				ControllingPlayer: landRegion.HomePlayer,
				HomePlayer:        landRegion.HomePlayer,
				Forest:            landRegion.Forest,
				Castle:            landRegion.Castle,
				Neighbors:         make([]gameboard.Neighbor, 0),
				IncomingMoves:     make([]gameboard.Order, 0),
				IncomingSupports:  make([]gameboard.Order, 0),
			}

			board.Regions[region.Name] = region
		}
	}

	for _, sea := range jsonBoard.Seas {
		region := gameboard.Region{
			Name:             sea.Name,
			Neighbors:        make([]gameboard.Neighbor, 0),
			IncomingMoves:    make([]gameboard.Order, 0),
			IncomingSupports: make([]gameboard.Order, 0),
		}

		board.Regions[region.Name] = region
	}

	for _, neighbor := range jsonBoard.Neighbors {
		region1, ok1 := board.Regions[neighbor.Region1]
		region2, ok2 := board.Regions[neighbor.Region2]

		if !ok1 || !ok2 {
			return gameboard.Board{}, fmt.Errorf(
				"error in board config: neighbor relation %s <-> %s",
				neighbor.Region1,
				neighbor.Region2,
			)
		}

		region1.Neighbors = append(region1.Neighbors, gameboard.Neighbor{
			Name:        region2.Name,
			AcrossWater: neighbor.River || (region1.Sea && !region2.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})

		region2.Neighbors = append(region2.Neighbors, gameboard.Neighbor{
			Name:        region1.Name,
			AcrossWater: neighbor.River || (region2.Sea && !region1.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})
	}

	return board, nil
}
