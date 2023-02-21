package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

// Calculates battle between a single attacker and an unconquered region.
// Sends the resulting battle to the given battleReceiver.
func calculateSingleplayerBattle(
	region gametypes.Region,
	move gametypes.Order,
	battleReceiver chan<- gametypes.Battle,
	messenger Messenger,
) {
	playerResults := map[string]gametypes.Result{
		move.Player: {Parts: attackModifiers(move, region, false, false, true), Move: move},
	}

	appendSupportMods(playerResults, region, false, messenger)

	battleReceiver <- gametypes.Battle{Results: calculateTotals(playerResults)}
}

// Calculates battle when attacked region is defended or has multiple attackers.
// Takes in parameter for whether to account for defender in battle (most often true).
// Sends the resulting battle to the given battleReceiver.
func calculateMultiplayerBattle(
	region gametypes.Region,
	includeDefender bool,
	battleReceiver chan<- gametypes.Battle,
	messenger Messenger,
) {
	playerResults := make(map[string]gametypes.Result)

	for _, move := range region.IncomingMoves {
		playerResults[move.Player] = gametypes.Result{
			Parts: attackModifiers(move, region, true, false, includeDefender),
			Move:  move,
		}
	}

	if !region.IsEmpty() && includeDefender {
		playerResults[region.Unit.Player] = gametypes.Result{
			Parts:          defenseModifiers(region),
			DefenderRegion: region.Name,
		}
	}

	appendSupportMods(playerResults, region, includeDefender, messenger)

	battleReceiver <- gametypes.Battle{Results: calculateTotals(playerResults)}
}

// Calculates battle when units from two regions attack each other simultaneously.
// Sends the resulting battle to the given battleReceiver.
func calculateBorderBattle(
	region1 gametypes.Region,
	region2 gametypes.Region,
	battleReceiver chan<- gametypes.Battle,
	messenger Messenger,
) {
	move1 := region1.Order
	move2 := region2.Order
	playerResults := map[string]gametypes.Result{
		move1.Player: {Parts: attackModifiers(move1, region2, true, true, false), Move: move1},
		move2.Player: {Parts: attackModifiers(move2, region1, true, true, false), Move: move2},
	}

	appendSupportMods(playerResults, region2, false, messenger)
	appendSupportMods(playerResults, region1, false, messenger)

	battleReceiver <- gametypes.Battle{Results: calculateTotals(playerResults)}
}

// Returns modifiers (including dice roll) of defending unit in the region.
// Assumes that the region is not empty.
func defenseModifiers(region gametypes.Region) []gametypes.Modifier {
	modifiers := []gametypes.Modifier{gametypes.RollDiceBonus()}

	if unitModifier, hasModifier := region.Unit.BattleModifier(false); hasModifier {
		modifiers = append(modifiers, unitModifier)
	}

	return modifiers
}

// Returns modifiers (including dice roll) of move order attacking a region.
// Other parameters affect which modifiers are added:
// otherAttackers for whether there are other moves involved in this battle,
// borderBattle for whether this is a battle between two moves moving against each other,
// includeDefender for whether a potential defending unit in the region should be included.
func attackModifiers(
	move gametypes.Order,
	region gametypes.Region,
	otherAttackers bool,
	borderBattle bool,
	includeDefender bool,
) []gametypes.Modifier {
	mods := []gametypes.Modifier{}

	neighbor, adjacent := region.GetNeighbor(move.Origin, move.Via)

	// Assumes danger zone checks have been made before battle,
	// and thus adds surprise modifier to attacker coming across such zones.
	if adjacent && neighbor.DangerZone != "" {
		mods = append(mods, gametypes.SurpriseAttackBonus())
	}

	// Terrain modifiers should be added if:
	// - Region is uncontrolled, and this unit is the only attacker.
	// - Destination is controlled and defended, and this is not a border conflict.
	if (!region.IsControlled() && !otherAttackers) ||
		(region.IsControlled() && !region.IsEmpty() && includeDefender && !borderBattle) {

		if region.Forest {
			mods = append(mods, gametypes.ForestAttackerPenalty())
		}

		if region.Castle {
			mods = append(mods, gametypes.CastleAttackerPenalty())
		}

		// If origin region is not adjacent to destination, the move is transported and takes water
		// penalty. Moves across rivers or from sea to land also take this penalty.
		if !adjacent || neighbor.AcrossWater {
			mods = append(mods, gametypes.AttackAcrossWaterPenalty())
		}
	}

	if unitModifier, hasModifier := region.Unit.BattleModifier(region.Castle); hasModifier {
		mods = append(mods, unitModifier)
	}

	mods = append(mods, gametypes.RollDiceBonus())

	return mods
}

// Calculates totals for the given map of player IDs to results, and returns them as a list.
func calculateTotals(playerResults map[string]gametypes.Result) []gametypes.Result {
	results := make([]gametypes.Result, 0)

	for _, result := range playerResults {
		total := 0
		for _, mod := range result.Parts {
			total += mod.Value
		}

		result.Total = total

		results = append(results, result)
	}

	return results
}
