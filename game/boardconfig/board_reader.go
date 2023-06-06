package boardconfig

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"hermannm.dev/bfh-server/game/gametypes"
)

//go:embed bfh_5players.json
var boardConfigFiles embed.FS

type JSONBoard struct {
	Name               string                      `json:"name"`
	WinningCastleCount int                         `json:"winningCastleCount"`
	Nations            map[string][]JSONLandRegion `json:"nations"`
	Seas               []JSONSeaRegion             `json:"seas"`
	Neighbors          []JSONNeighbor              `json:"neighbors"`
}

type JSONLandRegion struct {
	Name       string `json:"name"`
	Forest     bool   `json:"forest"`
	Castle     bool   `json:"castle"`
	HomePlayer string `json:"homePlayer"`
}

type JSONSeaRegion struct {
	Name string `json:"name"`
}

type JSONNeighbor struct {
	Region1    string `json:"region1"`
	Region2    string `json:"region2"`
	River      bool   `json:"river"`
	Cliffs     bool   `json:"cliffs"`
	DangerZone string `json:"dangerZone"`
}

func ReadBoardFromConfigFile(boardID string) (gametypes.Board, error) {
	content, err := boardConfigFiles.ReadFile(fmt.Sprintf("%s.json", boardID))
	if err != nil {
		return gametypes.Board{}, fmt.Errorf(
			"failed to read config file '%s.json': %w", boardID, err,
		)
	}

	var jsonBoard JSONBoard
	if err := json.Unmarshal(content, &jsonBoard); err != nil {
		return gametypes.Board{}, fmt.Errorf("failed to parse board config file: %w", err)
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
				IsForest:          landRegion.Forest,
				HasCastle:         landRegion.Castle,
			}

			board.Regions[region.Name] = region
		}
	}

	for _, sea := range jsonBoard.Seas {
		region := gametypes.Region{Name: sea.Name, IsSea: true}
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
			Name:          region2.Name,
			IsAcrossWater: neighbor.River || (region1.IsSea && !region2.IsSea),
			HasCliffs:     neighbor.Cliffs,
			DangerZone:    neighbor.DangerZone,
		})

		region2.Neighbors = append(region2.Neighbors, gametypes.Neighbor{
			Name:          region1.Name,
			IsAcrossWater: neighbor.River || (region2.IsSea && !region1.IsSea),
			HasCliffs:     neighbor.Cliffs,
			DangerZone:    neighbor.DangerZone,
		})
	}

	return board, nil
}

type PartialJSONBoard struct {
	Name               string `json:"name"`
	WinningCastleCount int    `json:"winningCastleCount"`
}

type BoardInfo struct {
	ID                 string `json:"id"`
	DescriptiveName    string `json:"descriptiveName"`
	WinningCastleCount int    `json:"winningCastleCount"`
}

func GetAvailableBoards() ([]BoardInfo, error) {
	directory, err := boardConfigFiles.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file directory: %w", err)
	}

	availableBoards := make([]BoardInfo, 0, len(directory))

	for _, directoryEntry := range directory {
		fullName := directoryEntry.Name()
		baseName, isJson := strings.CutSuffix(fullName, ".json")
		if !isJson {
			continue
		}

		content, err := boardConfigFiles.ReadFile(fullName)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to read config file '%s': %w", fullName, err,
			)
		}

		var board PartialJSONBoard
		if err := json.Unmarshal(content, &board); err != nil {
			return nil, fmt.Errorf("failed to parse board config file: %w", err)
		}

		boardInfo := BoardInfo{
			ID:                 baseName,
			DescriptiveName:    board.Name,
			WinningCastleCount: board.WinningCastleCount,
		}

		availableBoards = append(availableBoards, boardInfo)
	}

	return availableBoards, nil
}
