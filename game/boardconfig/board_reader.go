package boardconfig

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/gametypes"
)

// Embeds the JSON board config files from this folder.
//
//go:embed bfh_5players.json
var boardConfigFiles embed.FS

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
func ReadBoardFromConfigFile(boardConfigFileName string) (gametypes.Board, error) {
	content, err := boardConfigFiles.ReadFile(fmt.Sprintf("%s.json", boardConfigFileName))
	if err != nil {
		return gametypes.Board{}, fmt.Errorf(
			"failed to read config file '%s.json': %w", boardConfigFileName, err,
		)
	}

	var jsonBoard jsonBoard
	if err := json.Unmarshal(content, &jsonBoard); err != nil {
		return gametypes.Board{}, fmt.Errorf("failed to deserialize board config file: %w", err)
	}

	if jsonBoard.WinningCastleCount <= 0 {
		return gametypes.Board{}, errors.New("invalid winningCastleCount in board config")
	}

	board := gametypes.Board{
		Regions:            make(map[string]gametypes.Region),
		Name:               jsonBoard.Name,
		WinningCastleCount: jsonBoard.WinningCastleCount,
	}

	for nation, regions := range jsonBoard.Nations {
		for _, landRegion := range regions {
			region := gametypes.Region{
				Name:              landRegion.Name,
				Nation:            nation,
				ControllingPlayer: landRegion.HomePlayer,
				HomePlayer:        landRegion.HomePlayer,
				Forest:            landRegion.Forest,
				Castle:            landRegion.Castle,
			}

			board.Regions[region.Name] = region
		}
	}

	for _, sea := range jsonBoard.Seas {
		region := gametypes.Region{Name: sea.Name, Sea: true}
		board.Regions[region.Name] = region
	}

	for _, neighbor := range jsonBoard.Neighbors {
		region1, ok1 := board.Regions[neighbor.Region1]
		region2, ok2 := board.Regions[neighbor.Region2]

		if !ok1 || !ok2 {
			return gametypes.Board{}, fmt.Errorf(
				"failed to find regions for neighbor relation %s <-> %s in board config",
				neighbor.Region1,
				neighbor.Region2,
			)
		}

		region1.Neighbors = append(region1.Neighbors, gametypes.Neighbor{
			Name:        region2.Name,
			AcrossWater: neighbor.River || (region1.Sea && !region2.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})

		region2.Neighbors = append(region2.Neighbors, gametypes.Neighbor{
			Name:        region1.Name,
			AcrossWater: neighbor.River || (region2.Sea && !region1.Sea),
			Cliffs:      neighbor.Cliffs,
			DangerZone:  neighbor.DangerZone,
		})
	}

	return board, nil
}
