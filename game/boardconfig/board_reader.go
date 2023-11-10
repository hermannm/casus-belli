package boardconfig

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"hermannm.dev/bfh-server/game"
	"hermannm.dev/wrap"
)

//go:embed bfh_5players.json
var boardConfigFiles embed.FS

type JSONBoard struct {
	Name               string
	WinningCastleCount int
	Nations            map[string][]JSONLandRegion
	Seas               []JSONSeaRegion
	Neighbors          []JSONNeighbor
}

type JSONLandRegion struct {
	Name        string
	Forest      bool
	Castle      bool
	HomeFaction string
}

type JSONSeaRegion struct {
	Name string
}

type JSONNeighbor struct {
	Region1    string
	Region2    string
	River      bool
	Cliffs     bool
	DangerZone string
}

func ReadBoardFromConfigFile(
	boardID string,
) (board game.Board, name string, winningCastleCount int, err error) {
	content, err := boardConfigFiles.ReadFile(fmt.Sprintf("%s.json", boardID))
	if err != nil {
		return game.Board{}, "", 0, wrap.Errorf(
			err,
			"failed to read config file '%s.json'",
			boardID,
		)
	}

	var jsonBoard JSONBoard
	if err := json.Unmarshal(content, &jsonBoard); err != nil {
		return game.Board{}, "", 0, wrap.Error(err, "failed to parse board config file")
	}

	if jsonBoard.WinningCastleCount <= 0 {
		return game.Board{}, "", 0, errors.New("invalid winningCastleCount in board config")
	}

	board = make(game.Board)

	for nation, regions := range jsonBoard.Nations {
		for _, landRegion := range regions {
			region := game.Region{
				Name:               landRegion.Name,
				Nation:             nation,
				ControllingFaction: game.PlayerFaction(landRegion.HomeFaction),
				HomeFaction:        game.PlayerFaction(landRegion.HomeFaction),
				IsForest:           landRegion.Forest,
				HasCastle:          landRegion.Castle,
			}

			board[region.Name] = region
		}
	}

	for _, sea := range jsonBoard.Seas {
		region := game.Region{Name: sea.Name, IsSea: true}
		board[region.Name] = region
	}

	for _, neighbor := range jsonBoard.Neighbors {
		region1, ok1 := board[neighbor.Region1]
		region2, ok2 := board[neighbor.Region2]

		if !ok1 || !ok2 {
			return game.Board{}, "", 0, fmt.Errorf(
				"failed to find regions for neighbor relation '%s' <-> '%s' in board config",
				neighbor.Region1,
				neighbor.Region2,
			)
		}

		region1.Neighbors = append(
			region1.Neighbors,
			game.Neighbor{
				Name:          region2.Name,
				IsAcrossWater: neighbor.River || (region1.IsSea && !region2.IsSea),
				HasCliffs:     neighbor.Cliffs,
				DangerZone:    neighbor.DangerZone,
			},
		)

		region2.Neighbors = append(
			region2.Neighbors,
			game.Neighbor{
				Name:          region1.Name,
				IsAcrossWater: neighbor.River || (region2.IsSea && !region1.IsSea),
				HasCliffs:     neighbor.Cliffs,
				DangerZone:    neighbor.DangerZone,
			},
		)
	}

	return board, jsonBoard.Name, jsonBoard.WinningCastleCount, nil
}

type PartialJSONBoard struct {
	Name               string
	WinningCastleCount int
}

type BoardInfo struct {
	ID                 string
	DescriptiveName    string
	WinningCastleCount int
}

func GetAvailableBoards() ([]BoardInfo, error) {
	directory, err := boardConfigFiles.ReadDir(".")
	if err != nil {
		return nil, wrap.Error(err, "failed to read config file directory")
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
			return nil, wrap.Errorf(err, "failed to read config file '%s'", fullName)
		}

		var board PartialJSONBoard
		if err := json.Unmarshal(content, &board); err != nil {
			return nil, wrap.Error(err, "failed to parse board config file")
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
