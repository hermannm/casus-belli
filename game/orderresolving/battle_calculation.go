package orderresolving

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

func calculateSingleplayerBattle(
	region gametypes.Region,
	move gametypes.Order,
	battleReceiver chan<- gametypes.Battle,
	messenger Messenger,
) {
	results := map[gametypes.PlayerFaction]gametypes.Result{
		move.Faction: {Parts: attackModifiers(move, region, false, false, true), Move: move},
	}

	appendSupportModifiers(results, region, false, messenger)

	battleReceiver <- gametypes.Battle{Results: calculateTotals(results)}
}

func calculateMultiplayerBattle(
	region gametypes.Region,
	includeDefender bool,
	battleReceiver chan<- gametypes.Battle,
	messenger Messenger,
) {
	results := make(map[gametypes.PlayerFaction]gametypes.Result, len(region.IncomingMoves))

	for _, move := range region.IncomingMoves {
		results[move.Faction] = gametypes.Result{
			Parts: attackModifiers(move, region, true, false, includeDefender),
			Move:  move,
		}
	}

	if !region.IsEmpty() && includeDefender {
		results[region.Unit.Faction] = gametypes.Result{
			Parts:          defenseModifiers(region),
			DefenderRegion: region.Name,
		}
	}

	appendSupportModifiers(results, region, includeDefender, messenger)

	battleReceiver <- gametypes.Battle{Results: calculateTotals(results)}
}

// Battle where units from two regions attack each other simultaneously.
func calculateBorderBattle(
	region1 gametypes.Region,
	region2 gametypes.Region,
	battleReceiver chan<- gametypes.Battle,
	messenger Messenger,
) {
	move1 := region1.Order
	move2 := region2.Order
	results := map[gametypes.PlayerFaction]gametypes.Result{
		move1.Faction: {Parts: attackModifiers(move1, region2, true, true, false), Move: move1},
		move2.Faction: {Parts: attackModifiers(move2, region1, true, true, false), Move: move2},
	}

	appendSupportModifiers(results, region2, false, messenger)
	appendSupportModifiers(results, region1, false, messenger)

	battleReceiver <- gametypes.Battle{Results: calculateTotals(results)}
}

func defenseModifiers(region gametypes.Region) []gametypes.Modifier {
	modifiers := []gametypes.Modifier{gametypes.RollDiceBonus()}

	if unitModifier, hasModifier := region.Unit.BattleModifier(false); hasModifier {
		modifiers = append(modifiers, unitModifier)
	}

	return modifiers
}

func attackModifiers(
	move gametypes.Order,
	region gametypes.Region,
	hasOtherAttackers bool,
	isBorderBattle bool,
	includeDefender bool,
) []gametypes.Modifier {
	mods := []gametypes.Modifier{}

	neighbor, adjacent := region.GetNeighbor(move.Origin, move.ViaDangerZone)

	// Assumes danger zone checks have been made before battle,
	// and thus adds surprise modifier to attacker coming across such zones.
	if adjacent && neighbor.DangerZone != "" {
		mods = append(mods, gametypes.SurpriseAttackBonus())
	}

	isOnlyAttackerOnUncontrolledRegion := !region.IsControlled() && !hasOtherAttackers
	isAttackOnDefendedRegion := region.IsControlled() && !region.IsEmpty() && includeDefender &&
		!isBorderBattle
	includeTerrainModifiers := isOnlyAttackerOnUncontrolledRegion || isAttackOnDefendedRegion

	if includeTerrainModifiers {
		if region.IsForest {
			mods = append(mods, gametypes.ForestAttackerPenalty())
		}

		if region.HasCastle {
			mods = append(mods, gametypes.CastleAttackerPenalty())
		}

		isMovingAcrossWater := !adjacent || neighbor.IsAcrossWater
		if isMovingAcrossWater {
			mods = append(mods, gametypes.AttackAcrossWaterPenalty())
		}
	}

	if unitModifier, hasModifier := region.Unit.BattleModifier(region.HasCastle); hasModifier {
		mods = append(mods, unitModifier)
	}

	mods = append(mods, gametypes.RollDiceBonus())

	return mods
}

func calculateTotals(results map[gametypes.PlayerFaction]gametypes.Result) []gametypes.Result {
	resultList := make([]gametypes.Result, 0, len(results))

	for _, result := range results {
		total := 0
		for _, mod := range result.Parts {
			total += mod.Value
		}
		result.Total = total

		resultList = append(resultList, result)
	}

	return resultList
}
