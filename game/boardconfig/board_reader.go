package boardconfig

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/set"
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

func ReadBoardFromConfigFile(boardID string) (game.Board, game.BoardInfo, error) {
	content, err := boardConfigFiles.ReadFile(fmt.Sprintf("%s.json", boardID))
	if err != nil {
		return game.Board{}, game.BoardInfo{}, wrap.Errorf(
			err,
			"failed to read config file '%s.json'",
			boardID,
		)
	}

	var jsonBoard JSONBoard
	if err := json.Unmarshal(content, &jsonBoard); err != nil {
		return game.Board{}, game.BoardInfo{}, wrap.Error(err, "failed to parse board config file")
	}

	if jsonBoard.WinningCastleCount <= 0 {
		return game.Board{}, game.BoardInfo{}, errors.New(
			"invalid winningCastleCount in board config",
		)
	}

	board := make(game.Board)
	var factions set.ArraySet[game.PlayerFaction]

	for nation, regions := range jsonBoard.Nations {
		for _, landRegion := range regions {
			homeFaction := game.PlayerFaction(landRegion.HomeFaction)

			region := game.Region{
				Name:               game.RegionName(landRegion.Name),
				Nation:             nation,
				ControllingFaction: homeFaction,
				HomeFaction:        homeFaction,
				IsForest:           landRegion.Forest,
				HasCastle:          landRegion.Castle,
			}

			board[region.Name] = region

			if homeFaction != "" {
				factions.Add(homeFaction)
			}
		}
	}

	for _, sea := range jsonBoard.Seas {
		region := game.Region{Name: game.RegionName(sea.Name), IsSea: true}
		board[region.Name] = region
	}

	for _, neighbor := range jsonBoard.Neighbors {
		region1, ok1 := board[game.RegionName(neighbor.Region1)]
		region2, ok2 := board[game.RegionName(neighbor.Region2)]

		if !ok1 || !ok2 {
			return game.Board{}, game.BoardInfo{}, fmt.Errorf(
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
				DangerZone:    game.DangerZone(neighbor.DangerZone),
			},
		)

		region2.Neighbors = append(
			region2.Neighbors,
			game.Neighbor{
				Name:          region1.Name,
				IsAcrossWater: neighbor.River || (region2.IsSea && !region1.IsSea),
				HasCliffs:     neighbor.Cliffs,
				DangerZone:    game.DangerZone(neighbor.DangerZone),
			},
		)
	}

	if factions.Size() == 0 {
		return game.Board{}, game.BoardInfo{}, errors.New(
			"found no playable factions in board config",
		)
	}

	boardInfo := game.BoardInfo{
		ID:                 boardID,
		Name:               jsonBoard.Name,
		WinningCastleCount: jsonBoard.WinningCastleCount,
		PlayerFactions:     factions.ToSlice(),
	}

	return board, boardInfo, nil
}

type PartialJSONBoard struct {
	Name               string
	WinningCastleCount int
	Nations            map[string][]struct {
		HomeFaction string
	}
}

func GetAvailableBoards() ([]game.BoardInfo, error) {
	directory, err := boardConfigFiles.ReadDir(".")
	if err != nil {
		return nil, wrap.Error(err, "failed to read config file directory")
	}

	availableBoards := make([]game.BoardInfo, len(directory))
	var goroutines errgroup.Group

	for i, directoryEntry := range directory {
		i, directoryEntry := i, directoryEntry // Avoids mutating loop variables

		goroutines.Go(func() error {
			fullName := directoryEntry.Name()
			baseName, isJson := strings.CutSuffix(fullName, ".json")
			if !isJson {
				return errors.New("non-JSON board config file found")
			}

			file, err := boardConfigFiles.Open(fullName)
			if err != nil {
				return wrap.Errorf(err, "failed to read config file '%s'", fullName)
			}

			var board PartialJSONBoard
			if err := json.NewDecoder(file).Decode(&board); err != nil {
				return wrap.Errorf(err, "failed to parse board config file '%s'", fullName)
			}

			var factions set.ArraySet[game.PlayerFaction]
			for _, regions := range board.Nations {
				for _, region := range regions {
					if region.HomeFaction != "" {
						factions.Add(game.PlayerFaction(region.HomeFaction))
					}
				}
			}
			if factions.Size() == 0 {
				return fmt.Errorf("found no playable factions in board config file '%s'", fullName)
			}

			boardInfo := game.BoardInfo{
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
