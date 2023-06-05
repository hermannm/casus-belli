package gametypes

import (
	"hermannm.dev/set"
)

// Results of a battle between players, an attempt to conquer a neutral region, or an attempt to
// cross a danger zone.
type Battle struct {
	// If length is one, the battle was a neutral region conquest attempt or danger zone crossing.
	// If length is more than one, the battle was between players.
	Results []Result `json:"results"`

	// If battle was from a danger zone crossing: name of the danger zone, otherwise blank.
	DangerZone string `json:"dangerZone,omitempty"`
}

// Dice and modifier result for a battle.
type Result struct {
	Total int        `json:"total"`
	Parts []Modifier `json:"parts"`

	// If result of a move order to the battle: the move order in question, otherwise empty.
	Move Order `json:"move"`

	// If result of a defending unit in a region: the name of the region, otherwise blank.
	DefenderRegion string `json:"defenderRegion,omitempty"`
}

// Numbers to beat in different types of battles.
const (
	// Number to beat when attempting to conquer a neutral region.
	RequirementConquer int = 4

	// Number to beat when attempting to cross a danger zone.
	RequirementDangerZone int = 3
)

// Returns regions involved in the battle - typically 1, but 2 if it was a border battle.
func (battle Battle) RegionNames() []string {
	nameSet := set.New[string]()

	for _, result := range battle.Results {
		if result.DefenderRegion != "" {
			nameSet.Add(result.DefenderRegion)
		} else if result.Move.Destination != "" {
			nameSet.Add(result.Move.Destination)
		}
	}

	return nameSet.ToSlice()
}

func (battle Battle) IsBorderConflict() bool {
	return len(battle.Results) == 2 &&
		(battle.Results[0].Move.Destination == battle.Results[1].Move.Origin) &&
		(battle.Results[1].Move.Destination == battle.Results[0].Move.Origin)
}

// In case of a battle against an unconquered region or a danger zone, only one player is returned
// in one of the lists.
//
// In case of a battle between players, multiple winners are returned in the case of a tie.
func (battle Battle) WinnersAndLosers() (winners []string, losers []string) {
	if len(battle.Results) == 1 {
		result := battle.Results[0]

		if battle.DangerZone != "" {
			if result.Total >= RequirementDangerZone {
				return []string{result.Move.Player}, nil
			} else {
				return nil, []string{result.Move.Player}
			}
		}

		if result.Total >= RequirementConquer {
			return []string{result.Move.Player}, nil
		} else {
			return nil, []string{result.Move.Player}
		}
	}

	highestResult := 0
	for _, result := range battle.Results {
		if result.Total > highestResult {
			highestResult = result.Total
		}
	}

	for _, result := range battle.Results {
		if result.Total >= highestResult {
			winners = append(winners, result.Move.Player)
		} else {
			losers = append(losers, result.Move.Player)
		}
	}

	return winners, losers
}
