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
}

// Dice and modifier result for a battle.
type Result struct {
	Total int
	Parts []Modifier

	// If result of a move order to the battle: the move order in question, otherwise empty.
	Move Order

	// If result of a defending unit in a region: the faction of the defender, otherwise blank.
	DefenderFaction PlayerFaction `json:",omitempty"`
}

type ResultMap map[PlayerFaction]*Result

func (resultMap ResultMap) toBattle() Battle {
	battle := Battle{
		Results: make([]Result, 0, len(resultMap)),
	}
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

func (game *Game) calculateSingleplayerBattle(region *Region) {
	move := region.incomingMoves[0]
	resultMap := ResultMap{
		move.Faction: {Parts: attackModifiers(move, region, false, false), Move: move},
	}

	remainingSupports := resultMap.addAutomaticSupports(region, region.incomingMoves, false)
	if len(remainingSupports) == 0 {
		game.resolveSingleplayerBattle(resultMap.toBattle())
		return
	}

	region.resolving = true
	go func() {
		game.callSupportForRegion(region, remainingSupports, region.incomingMoves, resultMap, false)
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

	if !region.empty() {
		resultMap[region.Unit.Faction] = &Result{
			Parts:           defenseModifiers(region),
			DefenderFaction: region.Unit.Faction,
		}
	}

	remainingSupports := resultMap.addAutomaticSupports(region, region.incomingMoves, false)
	if len(remainingSupports) == 0 {
		game.resolveMultiplayerBattle(resultMap.toBattle())
		return
	}

	region.resolving = true
	go func() {
		game.callSupportForRegion(region, remainingSupports, region.incomingMoves, resultMap, false)
		game.battleReceiver <- resultMap.toBattle()
	}()
}

// Battle where units from two regions attack each other simultaneously.
func (game *Game) calculateBorderBattle(region1 *Region, region2 *Region) {
	moveToRegion1, moveToRegion2 := region2.order, region1.order
	movesToRegion1, movesToRegion2 := []Order{moveToRegion1}, []Order{moveToRegion2}

	resultMap := ResultMap{
		moveToRegion1.Faction: {
			Parts: attackModifiers(moveToRegion1, region1, true, true),
			Move:  moveToRegion1,
		},
		moveToRegion2.Faction: {
			Parts: attackModifiers(moveToRegion2, region2, true, true),
			Move:  moveToRegion2,
		},
	}

	remainingSupports1 := resultMap.addAutomaticSupports(region1, movesToRegion1, true)
	region1Done := len(remainingSupports1) == 0

	remainingSupports2 := resultMap.addAutomaticSupports(region2, movesToRegion2, true)
	region2Done := len(remainingSupports2) == 0

	if region1Done && region2Done {
		game.resolveBorderBattle(resultMap.toBattle())
		return
	}

	region1.resolving = true
	region2.resolving = true
	go func() {
		callSupportRegion1 := func() {
			game.callSupportForRegion(region1, remainingSupports1, movesToRegion1, resultMap, true)
		}
		callSupportRegion2 := func() {
			game.callSupportForRegion(region2, remainingSupports2, movesToRegion2, resultMap, true)
		}

		if !region1Done && !region2Done {
			var waitGroup sync.WaitGroup
			waitGroup.Add(2)
			go func() {
				callSupportRegion1()
				waitGroup.Done()
			}()
			go func() {
				callSupportRegion2()
				waitGroup.Done()
			}()
			waitGroup.Wait()
		} else if !region1Done {
			callSupportRegion1()
		} else if !region2Done {
			callSupportRegion2()
		}

		game.battleReceiver <- resultMap.toBattle()
	}()
}

// Adds modifiers for support orders from players involved in the battle, as we assume they always
// want to support themselves, so we don't need to ask. Returns support orders that could not be
// added automatically.
func (resultMap ResultMap) addAutomaticSupports(
	region *Region,
	incomingMoves []Order,
	borderBattle bool,
) (remainingSupports []Order) {
SupportLoop:
	for _, support := range region.incomingSupports {
		if !region.empty() && !borderBattle && support.Faction == region.Unit.Faction {
			resultMap.addSupport(support.Faction, support.Faction)
			continue
		}

		for _, move := range incomingMoves {
			if move.Faction == support.Faction {
				resultMap.addSupport(support.Faction, support.Faction)
				continue SupportLoop
			}
		}

		remainingSupports = append(remainingSupports, support)
	}

	return remainingSupports
}

// Calls support from support orders to the given region, and adds modifiers to the result map.
func (game *Game) callSupportForRegion(
	region *Region,
	supports []Order,
	incomingMoves []Order,
	resultMap ResultMap,
	borderBattle bool,
) {
	if len(supports) == 1 {
		support := supports[0]
		supported := game.callSupportFromPlayer(support, region, incomingMoves, borderBattle)
		if supported != "" {
			resultMap.addSupport(support.Faction, supported)
		}
		return
	}

	var resultsLock sync.Mutex
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(supports))

	for _, support := range supports {
		support := support // Avoids mutating loop variable

		go func() {
			supported := game.callSupportFromPlayer(support, region, incomingMoves, borderBattle)
			if supported != "" {
				resultsLock.Lock()
				resultMap.addSupport(support.Faction, supported)
				resultsLock.Unlock()
			}
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()
}

func (game *Game) callSupportFromPlayer(
	support Order,
	region *Region,
	incomingMoves []Order,
	borderBattle bool,
) (supported PlayerFaction) {
	supportableFactions := make([]PlayerFaction, 0, len(incomingMoves)+1)
	for _, move := range incomingMoves {
		supportableFactions = append(supportableFactions, move.Faction)
	}
	if !region.empty() && !borderBattle {
		supportableFactions = append(supportableFactions, region.Unit.Faction)
	}

	if err := game.messenger.SendSupportRequest(
		support.Faction,
		support.Origin,
		region.Name,
		supportableFactions,
	); err != nil {
		game.log.ErrorCause(err, "failed to send request for support", support.logAttribute())
		return ""
	}

	supported, err := game.messenger.AwaitSupport(support.Faction, support.Origin, region.Name)
	if err != nil {
		game.log.ErrorCause(err, "failed to receive support declaration", support.logAttribute())
		return ""
	}

	if supported != "" && !slices.Contains(supportableFactions, supported) {
		err := fmt.Errorf("received invalid supported faction '%s'", supported)
		game.log.Error(err, support.logAttribute())
		game.messenger.SendError(support.Faction, err)
		return ""
	}

	return supported
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
		game.board[region].resolving = false
	}
}

func (game *Game) resolveSingleplayerBattle(battle Battle) {
	winners, _ := battle.winnersAndLosers()
	move := battle.Results[0].Move

	if len(winners) != 1 {
		game.board.retreatMove(move)
		return
	}

	game.board.succeedMove(move)
	if err := game.messenger.SendBattleResults(battle); err != nil {
		game.log.Error(err)
	}
}

func (game *Game) resolveMultiplayerBattle(battle Battle) {
	winners, losers := battle.winnersAndLosers()
	tie := len(winners) > 1

	for _, result := range battle.Results {
		if result.DefenderFaction != "" {
			// If the defender won or or was part of a tie, nothing changes for them.
			// If an attacker won alone and the defender controlled the region, the defender will be
			// removed as part of succeedMove for the winner.
			// If the defender was on the losing end of a tie in a battle with multiple combatants,
			// or the defender lost but did not control the region, we have to remove the unit here.
			if slices.Contains(losers, result.DefenderFaction) {
				// Guaranteed to have 1 element, since this is not a border battle
				regionName := battle.regionNames()[0]
				region := game.board[regionName]
				if tie {
					region.removeUnit()
				} else if !region.controlled() {
					region.removeUnit()
				}
			}
			continue
		}

		move := result.Move
		if slices.Contains(losers, move.Faction) {
			game.board.killMove(move)
			continue
		}

		if tie {
			game.board.retreatMove(move)
			continue
		}

		// If the destination is not controlled, then the winner will have to battle there before we
		// can succeed the move
		if game.board[move.Destination].controlled() {
			game.board.succeedMove(move)
		}
	}

	if err := game.messenger.SendBattleResults(battle); err != nil {
		game.log.Error(err)
	}
}

func (game *Game) resolveBorderBattle(battle Battle) {
	winners, losers := battle.winnersAndLosers()
	move1 := battle.Results[0].Move
	move2 := battle.Results[1].Move

	// If battle was a tie, both moves retreat
	if len(winners) > 1 {
		// Remove both orders before retreating, so they don't think their origins are attacked
		game.board.removeOrder(move1)
		game.board.removeOrder(move2)

		game.board.retreatMove(move1)
		game.board.retreatMove(move2)
		return
	}

	loser := losers[0]
	for _, move := range []Order{move1, move2} {
		// Only the loser is affected by the results of the border battle; the winner may still have
		// to win a battle in the destination region, which will be handled by the next cycle of the
		// move resolver
		if move.Faction == loser {
			game.board.killMove(move)
			break
		}
	}

	if err := game.messenger.SendBattleResults(battle); err != nil {
		game.log.Error(err)
	}
}

func (battle Battle) isBorderBattle() bool {
	return len(battle.Results) == 2 &&
		(battle.Results[0].Move.Destination == battle.Results[1].Move.Origin) &&
		(battle.Results[1].Move.Destination == battle.Results[0].Move.Origin)
}

// Number to beat when attempting to conquer a neutral region.
const MinDiceResultToConquerNeutralRegion int = 4

// In case of a battle against an unconquered region or a danger zone, only one player faction is
// returned in one of the lists.
//
// In case of a battle between players, multiple winners are returned in the case of a tie.
func (battle Battle) winnersAndLosers() (winners []PlayerFaction, losers []PlayerFaction) {
	if len(battle.Results) == 1 {
		result := battle.Results[0]

		if result.Total >= MinDiceResultToConquerNeutralRegion {
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
		var faction PlayerFaction
		if result.DefenderFaction != "" {
			faction = result.DefenderFaction
		} else {
			faction = result.Move.Faction
		}

		if result.Total >= highestResult {
			winners = append(winners, faction)
		} else {
			losers = append(losers, faction)
		}
	}

	return winners, losers
}

// Returns regions involved in the battle - typically 1, but 2 if it was a border battle.
func (battle Battle) regionNames() []RegionName {
	nameSet := set.ArraySetWithCapacity[RegionName](2)

	for _, result := range battle.Results {
		if !result.Move.isNone() {
			nameSet.Add(result.Move.Destination)
		}
	}

	return nameSet.ToSlice()
}
