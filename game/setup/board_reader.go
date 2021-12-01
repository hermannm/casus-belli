package setup

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/immerse-ntnu/despot-dash/server/game"
)

// Utility type for json unmarshaling.
type board struct {
	Areas     []area     `json:"areas"`
	Neighbors []neighbor `json:"neighbors"`
}

// Utility type for json unmarshaling.
type area struct {
	Name   string `json:"name"`
	Sea    bool   `json:"sea"`
	Forest bool   `json:"forest"`
	Castle bool   `json:"castle"`
	Home   string `json:"home"`
}

// Utility type for json unmarshaling.
type neighbor struct {
	Area1      string `json:"area1"`
	Area2      string `json:"area2"`
	River      bool   `json:"river"`
	Cliffs     bool   `json:"cliffs"`
	DangerZone string `json:"dangerZone"`
}

// Reads and constructs the board matching the given number of players.
func ReadBoard(players int) (game.Board, error) {
	content, err := os.ReadFile(fmt.Sprintf("./game/setup/board_%dplayers.json", players))
	if err != nil {
		return nil, err
	}

	var jsonBoard board

	err = json.Unmarshal(content, &jsonBoard)
	if err != nil {
		return nil, err
	}

	board := make(game.Board)

	for _, jsonArea := range jsonBoard.Areas {
		area := game.BoardArea{
			Name:             jsonArea.Name,
			Control:          game.Player(jsonArea.Home),
			Home:             game.Player(jsonArea.Home),
			Sea:              jsonArea.Sea,
			Forest:           jsonArea.Forest,
			Castle:           jsonArea.Castle,
			Neighbors:        make([]game.Neighbor, 0),
			IncomingMoves:    make([]*game.Order, 0),
			IncomingSupports: make([]*game.Order, 0),
			Combats:          make([]game.Combat, 0),
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
