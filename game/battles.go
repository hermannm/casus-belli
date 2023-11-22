package game

import (
	"fmt"
	"slices"
	"sync"

	"hermannm.dev/set"
)

// Results of a battle between players, an attempt to conquer a neutral region, or an attempt to
// cross a danger zone.
type Battle struct {
	// If length is one, the battle was a neutral region conquest attempt or danger zone crossing.
	// If length is more than one, the battle was between players.
	Results []Result

	// If battle was from a danger zone crossing: name of the danger zone, otherwise blank.
	DangerZone DangerZone `json:",omitempty"`
}

// Dice and modifier result for a battle.
type Result struct {
	Total int
	Parts []Modifier

	// If result of a move order to the battle: the move order in question, otherwise empty.
	Move Order

	// If result of a defending unit in a region: the name of the region, otherwise blank.
	DefenderRegion RegionName `json:",omitempty"`
}

// Numbers to beat in different types of battles.
const (
	// Number to beat when attempting to conquer a neutral region.
	RequirementConquer int = 4

	// Number to beat when attempting to cross a danger zone.
	RequirementDangerZone int = 3
)

type ResultMap map[PlayerFaction]*Result

func (resultMap ResultMap) moves(destination RegionName) []Order {
	var moves []Order
	for _, result := range resultMap {
		if result.DefenderRegion == "" && result.Move.Destination == destination {
			moves = append(moves, result.Move)
		}
	}
	return moves
}

func (resultMap ResultMap) toBattle() Battle {
	battle := Battle{Results: make([]Result, 0, len(resultMap))}
	for _, result := range resultMap {
		for _, mod := range result.Parts {
			result.Total += mod.Value
		}
		battle.Results = append(battle.Results, *result)
	}
	return battle
}

func (resultMap ResultMap) addSupport(from PlayerFaction, to PlayerFaction) {
	resultMap[to].Parts = append(
		resultMap[to].Parts,
		Modifier{Type: ModifierSupport, Value: 1, SupportingFaction: from},
	)
}

func (game *Game) calculateSingleplayerBattle(region *Region, move Order) {
	resultMap := ResultMap{
		move.Faction: {Parts: attackModifiers(move, region, false, false), Move: move},
	}

	if !region.isSupported() {
		game.resolveSingleplayerBattle(resultMap.toBattle())
		return
	}

	region.resolving = true
	go func() {
		game.callSupportForRegion(region, resultMap, false)
		game.battleReceiver <- resultMap.toBattle()
	}()
}

func (game *Game) calculateMultiplayerBattle(region *Region) {
	resultMap := make(ResultMap, len(region.incomingMoves))

	for _, move := range region.incomingMoves {
		resultMap[move.Faction] = &Result{
			Parts: attackModifiers(move, region, true, false),
			Move:  move,
		}
	}

	if !region.isEmpty() {
		resultMap[region.Unit.Faction] = &Result{
			Parts:          defenseModifiers(region),
			DefenderRegion: region.Name,
		}
	}

	if !region.isSupported() {
		game.resolveMultiplayerBattle(resultMap.toBattle())
		return
	}

	region.resolving = true
	go func() {
		game.callSupportForRegion(region, resultMap, false)
		game.battleReceiver <- resultMap.toBattle()
	}()
}

// Battle where units from two regions attack each other simultaneously.
func (game *Game) calculateBorderBattle(region1 *Region, region2 *Region) {
	move1 := region1.order
	move2 := region2.order
	resultMap := ResultMap{
		move1.Faction: {Parts: attackModifiers(move1, region2, true, true), Move: move1},
		move2.Faction: {Parts: attackModifiers(move2, region1, true, true), Move: move2},
	}

	if !region1.isSupported() && !region2.isSupported() {
		game.resolveBorderBattle(resultMap.toBattle())
		return
	}

	region1.resolving = true
	region2.resolving = true
	go func() {
		// TODO: make these two calls run concurrently (possibly race condition on results)
		game.callSupportForRegion(region1, resultMap, true)
		game.callSupportForRegion(region2, resultMap, true)

		game.battleReceiver <- resultMap.toBattle()
	}()
}

type supportDeclaration struct {
	from PlayerFaction
	to   PlayerFaction // Blank if nobody were supported.
}

// Calls support from support orders to the given region, and adds modifiers to the result map.
func (game *Game) callSupportForRegion(
	region *Region,
	resultMap ResultMap,
	isBorderBattle bool,
) {
	supports := region.incomingSupports
	supportReceiver := make(chan supportDeclaration, len(supports))

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(supports))

	incomingMoves := []Order{}
	for _, result := range resultMap {
		if result.DefenderRegion != "" {
			continue
		}
		if result.Move.Destination == region.Name {
			incomingMoves = append(incomingMoves, result.Move)
		}
	}

	for _, support := range supports {
		go game.callSupportFromPlayer(
			support,
			region,
			incomingMoves,
			isBorderBattle,
			supportReceiver,
			&waitGroup,
		)
	}

	waitGroup.Wait()
	close(supportReceiver)

	for support := range supportReceiver {
		if support.to == "" {
			continue
		}

		resultMap.addSupport(support.from, support.to)
	}
}

func (game *Game) callSupportFromPlayer(
	support Order,
	region *Region,
	moves []Order,
	isBorderBattle bool,
	supportReceiver chan<- supportDeclaration,
	waitGroup *sync.WaitGroup,
) {
	defer waitGroup.Done()

	includeDefender := !region.isEmpty() && !isBorderBattle
	if includeDefender && region.Unit.Faction == support.Faction {
		supportReceiver <- supportDeclaration{from: support.Faction, to: support.Faction}
		return
	}

	for _, move := range moves {
		if support.Faction == move.Faction {
			supportReceiver <- supportDeclaration{from: support.Faction, to: support.Faction}
			return
		}
	}

	supportableFactions := make([]PlayerFaction, 0, len(moves)+1)
	for _, move := range moves {
		supportableFactions = append(supportableFactions, move.Faction)
	}
	if includeDefender {
		supportableFactions = append(supportableFactions, region.Unit.Faction)
	}

	if err := game.messenger.SendSupportRequest(
		support.Faction,
		support.Origin,
		region.Name,
		supportableFactions,
	); err != nil {
		game.log.ErrorCause(err, "failed to send support request")
		supportReceiver <- supportDeclaration{from: support.Faction, to: ""}
		return
	}

	supported, err := game.messenger.AwaitSupport(support.Faction, support.Origin, region.Name)
	if err != nil {
		game.log.ErrorCausef(
			err,
			"failed to receive support declaration from faction '%s' in '%s' for battle in '%s'",
			support.Faction,
			support.Origin,
			region.Name,
		)
		supportReceiver <- supportDeclaration{from: support.Faction, to: ""}
		return
	}

	if supported != "" && !slices.Contains(supportableFactions, supported) {
		err := fmt.Errorf(
			"received invalid supported faction '%s' from support order in '%s' for battle in '%s'",
			supported,
			support.Origin,
			region.Name,
		)
		game.log.Error(err)
		game.messenger.SendError(support.Faction, err)
		supportReceiver <- supportDeclaration{from: support.Faction, to: ""}
		return
	}

	supportReceiver <- supportDeclaration{from: support.Faction, to: supported}
}

func (game *Game) resolveBattle(battle Battle) {
	if battle.isBorderBattle() {
		game.resolveBorderBattle(battle)
	} else if len(battle.Results) == 1 {
		game.resolveSingleplayerBattle(battle)
	} else {
		game.resolveMultiplayerBattle(battle)
	}

	for _, region := range battle.regionNames() {
		game.Board[region].resolving = false
	}
}

func (game *Game) resolveSingleplayerBattle(battle Battle) {
	winners, _ := battle.winnersAndLosers()
	move := battle.Results[0].Move

	if len(winners) != 1 {
		game.Board.removeOrder(move)
		game.attemptRetreat(move)
		return
	}

	game.succeedMove(move)
	game.resolvedBattles = append(game.resolvedBattles, battle)
	game.messenger.SendBattleResults(battle)
}

func (game *Game) resolveMultiplayerBattle(battle Battle) {
	winners, losers := battle.winnersAndLosers()
	tie := len(winners) > 1

	for _, result := range battle.Results {
		if result.DefenderRegion != "" {
			// If the defender won or or was part of a tie, nothing changes for them.
			// If an attacker won alone and the defender controlled the region, the defender will be
			// removed as part of succeedMove for the winner.
			// If the defender was on the losing end of a tie in a battle with multiple combatants,
			// or the defender lost but did not control the region, we have to remove the unit here.
			region := game.Board[result.DefenderRegion]
			if slices.Contains(losers, region.Unit.Faction) {
				if tie {
					region.removeUnit()
				} else if !region.isControlled() {
					region.removeUnit()
				}
			}
			continue
		}

		move := result.Move
		if slices.Contains(losers, move.Faction) {
			game.Board.removeOrder(move)
			game.Board[move.Origin].removeUnit()
			continue
		}

		if tie {
			game.Board.removeOrder(move)
			game.attemptRetreat(move)
			continue
		}

		// If the destination is not controlled, then the winner will have to battle there
		// before we can succeed the move
		if game.Board[move.Destination].isControlled() {
			game.succeedMove(move)
		}
	}

	game.resolvedBattles = append(game.resolvedBattles, battle)
	game.messenger.SendBattleResults(battle)
}

func (game *Game) resolveBorderBattle(battle Battle) {
	winners, losers := battle.winnersAndLosers()
	move1 := battle.Results[0].Move
	move2 := battle.Results[1].Move

	// If battle was a tie, both moves retreat
	if len(winners) > 1 {
		game.Board.removeOrder(move1)
		game.Board.removeOrder(move2)

		game.attemptRetreat(move1)
		game.attemptRetreat(move2)

		return
	}

	loser := losers[0]

	for _, move := range []Order{move1, move2} {
		// Only the loser is affected by the results of the border battle; the winner may still have
		// to win a battle in the destination region, which will be handled by the next cycle of the
		// move resolver
		if move.Faction == loser {
			game.Board.removeOrder(move)
			game.Board[move.Origin].removeUnit()
		}
	}

	game.resolvedBattles = append(game.resolvedBattles, battle)
	game.messenger.SendBattleResults(battle)
}

func (battle Battle) isBorderBattle() bool {
	return len(battle.Results) == 2 &&
		(battle.Results[0].Move.Destination == battle.Results[1].Move.Origin) &&
		(battle.Results[1].Move.Destination == battle.Results[0].Move.Origin)
}

// In case of a battle against an unconquered region or a danger zone, only one player faction is
// returned in one of the lists.
//
// In case of a battle between players, multiple winners are returned in the case of a tie.
func (battle Battle) winnersAndLosers() (winners []PlayerFaction, losers []PlayerFaction) {
	if len(battle.Results) == 1 {
		result := battle.Results[0]

		if battle.DangerZone != "" {
			if result.Total >= RequirementDangerZone {
				return []PlayerFaction{result.Move.Faction}, nil
			} else {
				return nil, []PlayerFaction{result.Move.Faction}
			}
		}

		if result.Total >= RequirementConquer {
			return []PlayerFaction{result.Move.Faction}, nil
		} else {
			return nil, []PlayerFaction{result.Move.Faction}
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
			winners = append(winners, result.Move.Faction)
		} else {
			losers = append(losers, result.Move.Faction)
		}
	}

	return winners, losers
}

// Returns regions involved in the battle - typically 1, but 2 if it was a border battle.
func (battle Battle) regionNames() []RegionName {
	nameSet := set.ArraySetWithCapacity[RegionName](2)

	for _, result := range battle.Results {
		if result.DefenderRegion != "" {
			nameSet.Add(result.DefenderRegion)
		} else if result.Move.Destination != "" {
			nameSet.Add(result.Move.Destination)
		}
	}

	return nameSet.ToSlice()
}

func (game *Game) attemptRetreat(move Order) {
	origin := game.Board[move.Origin]

	if origin.Unit == move.Unit {
		return
	}

	if origin.isAttacked() {
		origin.retreat = move
		return
	}

	if origin.isEmpty() {
		origin.Unit = move.Unit
	}
	origin.resolved = true
}
