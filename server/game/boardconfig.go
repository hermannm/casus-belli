package game

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
	"hermannm.dev/set"
	"hermannm.dev/wrap"
)

//go:embed boardconfig
var boardConfigFiles embed.FS

type boardConfig struct {
	Name               string                        `json:"name"`
	WinningCastleCount int                           `json:"winningCastleCount"`
	Nations            map[string][]landRegionConfig `json:"nations"`
	Seas               []seaRegionConfig             `json:"seas"`
	Neighbors          []neighborConfig              `json:"neighbors"`
}

type landRegionConfig struct {
	Name        string `json:"name"`
	Forest      bool   `json:"forest"`
	Castle      bool   `json:"castle"`
	HomeFaction string `json:"homeFaction"`
}

type seaRegionConfig struct {
	Name string `json:"name"`
}

type neighborConfig struct {
	Region1    string `json:"region1"`
	Region2    string `json:"region2"`
	River      bool   `json:"river"`
	Cliffs     bool   `json:"cliffs"`
	DangerZone string `json:"dangerZone"`
}

func ReadBoardFromConfigFile(boardID string) (Board, BoardInfo, error) {
	content, err := boardConfigFiles.ReadFile("boardconfig/" + boardID + ".json")
	if err != nil {
		return Board{}, BoardInfo{}, wrap.Errorf(
			err,
			"failed to read config file '%s.json'",
			boardID,
		)
	}

	var boardConfig boardConfig
	if err := json.Unmarshal(content, &boardConfig); err != nil {
		return Board{}, BoardInfo{}, wrap.Error(err, "failed to parse board config file")
	}

	if boardConfig.WinningCastleCount <= 0 {
		return Board{}, BoardInfo{}, errors.New(
			"invalid winningCastleCount in board config",
		)
	}

	board := make(Board)
	var factions set.ArraySet[PlayerFaction]

	for nation, regions := range boardConfig.Nations {
		for _, landRegion := range regions {
			homeFaction := PlayerFaction(landRegion.HomeFaction)

			region := Region{
				Name:               RegionName(landRegion.Name),
				Nation:             nation,
				ControllingFaction: homeFaction,
				HomeFaction:        homeFaction,
				Forest:             landRegion.Forest,
				Castle:             landRegion.Castle,
			}

			board[region.Name] = &region

			if homeFaction != "" {
				factions.Add(homeFaction)
			}
		}
	}

	for _, sea := range boardConfig.Seas {
		region := Region{Name: RegionName(sea.Name), Sea: true}
		board[region.Name] = &region
	}

	for _, neighbor := range boardConfig.Neighbors {
		region1, ok1 := board[RegionName(neighbor.Region1)]
		region2, ok2 := board[RegionName(neighbor.Region2)]

		if !ok1 || !ok2 {
			return Board{}, BoardInfo{}, fmt.Errorf(
				"failed to find regions for neighbor relation '%s' <-> '%s' in board config",
				neighbor.Region1,
				neighbor.Region2,
			)
		}

		region1.Neighbors = append(
			region1.Neighbors,
			Neighbor{
				Name:        region2.Name,
				AcrossWater: neighbor.River || (region1.Sea && !region2.Sea),
				Cliffs:      neighbor.Cliffs,
				DangerZone:  DangerZone(neighbor.DangerZone),
			},
		)

		region2.Neighbors = append(
			region2.Neighbors,
			Neighbor{
				Name:        region1.Name,
				AcrossWater: neighbor.River || (region2.Sea && !region1.Sea),
				Cliffs:      neighbor.Cliffs,
				DangerZone:  DangerZone(neighbor.DangerZone),
			},
		)
	}

	if factions.Size() == 0 {
		return Board{}, BoardInfo{}, errors.New(
			"found no playable factions in board config",
		)
	}

	boardInfo := BoardInfo{
		ID:                 boardID,
		Name:               boardConfig.Name,
		WinningCastleCount: boardConfig.WinningCastleCount,
		PlayerFactions:     factions.ToSlice(),
	}

	return board, boardInfo, nil
}

type partialBoardConfig struct {
	Name               string `json:"name"`
	WinningCastleCount int    `json:"winningCastleCount"`
	Nations            map[string][]struct {
		HomeFaction string `json:"homeFaction"`
	} `json:"nations"`
}

func GetAvailableBoards() ([]BoardInfo, error) {
	directory, err := boardConfigFiles.ReadDir("boardconfig")
	if err != nil {
		return nil, wrap.Error(err, "failed to read config file directory")
	}

	availableBoards := make([]BoardInfo, len(directory))
	var goroutines errgroup.Group

	for i, directoryEntry := range directory {
		goroutines.Go(func() error {
			fullName := directoryEntry.Name()
			baseName, isJson := strings.CutSuffix(fullName, ".json")
			if !isJson {
				return errors.New("non-JSON board config file found")
			}

			file, err := boardConfigFiles.Open("boardconfig/" + fullName)
			if err != nil {
				return wrap.Errorf(err, "failed to read config file '%s'", fullName)
			}

			var board partialBoardConfig
			if err := json.NewDecoder(file).Decode(&board); err != nil {
				return wrap.Errorf(err, "failed to parse board config file '%s'", fullName)
			}

			var factions set.ArraySet[PlayerFaction]
			for _, regions := range board.Nations {
				for _, region := range regions {
					if region.HomeFaction != "" {
						factions.Add(PlayerFaction(region.HomeFaction))
					}
				}
			}
			if factions.Size() == 0 {
				return fmt.Errorf("found no playable factions in board config file '%s'", fullName)
			}

			boardInfo := BoardInfo{
				ID:                 baseName,
				Name:               board.Name,
				WinningCastleCount: board.WinningCastleCount,
				PlayerFactions:     factions.ToSlice(),
			}

			availableBoards[i] = boardInfo
			return nil
		})
	}

	if err := goroutines.Wait(); err != nil {
		return nil, err
	}

	return availableBoards, nil
}
