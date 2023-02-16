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
		Regions:            make(map[string]board.Region),
		Name:               jsonBoard.Name,
		WinningCastleCount: jsonBoard.WinningCastleCount,
	}

	for nation, regions := range jsonBoard.Nations {
		for _, landRegion := range regions {
			region := board.Region{
				Name:              landRegion.Name,
				Nation:            nation,
				ControllingPlayer: landRegion.HomePlayer,
				HomePlayer:        landRegion.HomePlayer,
				Forest:            landRegion.Forest,
				Castle:            landRegion.Castle,
				Neighbors:         make([]board.Neighbor, 0),
				IncomingMoves:     make([]board.Order, 0),
				IncomingSupports:  make([]board.Order, 0),
			}

			brd.Regions[region.Name] = region
		}
	}

	for _, sea := range jsonBoard.Seas {
		region := board.Region{
			Name:             sea.Name,
			Neighbors:        make([]board.Neighbor, 0),
			IncomingMoves:    make([]board.Order, 0),
			IncomingSupports: make([]board.Order, 0),
		}

		brd.Regions[region.Name] = region
	}

	for _, neighbor := range jsonBoard.Neighbors {
		region1, ok1 := brd.Regions[neighbor.Region1]
		region2, ok2 := brd.Regions[neighbor.Region2]

		if !ok1 || !ok2 {
			return board.Board{}, fmt.Errorf(
				"error in board config: neighbor relation %s <-> %s",
				neighbor.Region1,
				neighbor.Region2,
			)
		}

		region1.Neighbors = append(region1.Neighbors, board.Neighbor{
			Name:        region2.Name,
			AcrossWater: neighbor.River || (region1.Sea && !region2.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})

		region2.Neighbors = append(region2.Neighbors, board.Neighbor{
			Name:        region1.Name,
			AcrossWater: neighbor.River || (region2.Sea && !region1.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})
	}

	return brd, nil
}
