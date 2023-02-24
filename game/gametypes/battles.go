package gametypes

import "hermannm.dev/set"

// Results of a battle from conflicting move orders, an attempt to conquer a neutral region,
// or an attempt to cross a danger zone.
type Battle struct {
	// The dice and modifier results of the battle.
	// If length is one, the battle was a neutral conquer attempt.
	// If length is more than one, the battle was between players.
	Results []Result `json:"results"`

	// In case of danger zone crossing: name of the danger zone.
	DangerZone string `json:"dangerZone"`
}

// Dice and modifier result for a battle.
type Result struct {
	// The sum of the dice roll and modifiers.
	Total int `json:"total"`

	// The modifiers comprising the result, including the dice roll.
	Parts []Modifier `json:"parts"`

	// If result of a move order to the battle: the move order in question.
	Move Order `json:"move"`

	// If result of a defending unit in a region: the name of the region.
	DefenderRegion string `json:"defenderRegion"`
}

// Numbers to beat in different types of battles.
const (
	// Number to beat when attempting to conquer a neutral region.
	RequirementConquer int = 4

	// Number to beat when attempting to cross a danger zone.
	RequirementDangerZone int = 3
)

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

// Returns whether the battle was between two moves moving against each other.
func (battle Battle) IsBorderConflict() bool {
	return len(battle.Results) == 2 &&
		(battle.Results[0].Move.Destination == battle.Results[1].Move.Origin) &&
		(battle.Results[1].Move.Destination == battle.Results[0].Move.Origin)
}

// Goes through the results of the battle and finds the winners and losers.
//
// In case of a battle against an unconquered region or a danger zone, only one player is returned
// in one of the lists.
//
// In case of a battle between players, multiple winners are returned in the case of a tie.
func (battle Battle) WinnersAndLosers() (winners []string, losers []string) {
	// Checks if the battle was against an unconquered region.
	if len(battle.Results) == 1 {
		result := battle.Results[0]

		// Checks that order meets the requirement for crossing the danger zone.
		if battle.DangerZone != "" {
			if result.Total >= RequirementDangerZone {
				return []string{result.Move.Player}, nil
			} else {
				return nil, []string{result.Move.Player}
			}
		}

		// Checks that order meets the requirement for conquering the neutral region.
		if result.Total >= RequirementConquer {
			return []string{result.Move.Player}, nil
		} else {
			return nil, []string{result.Move.Player}
		}
	}

	// Finds the highest result among the players in the battle.
	highestResult := 0
	for _, result := range battle.Results {
		if result.Total > highestResult {
			highestResult = result.Total
		}
	}

	// Sorts combatants based on whether they exceeded the highest result.
	for _, result := range battle.Results {
		if result.Total >= highestResult {
			winners = append(winners, result.Move.Player)
		} else {
			losers = append(losers, result.Move.Player)
		}
	}

	return winners, losers
}
