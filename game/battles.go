package game

import (
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

func (game *Game) calculateSingleplayerBattle(region Region, move Order) {
	results := map[PlayerFaction]Result{
		move.Faction: {Parts: attackModifiers(move, region, false, false, true), Move: move},
	}

	game.callSupportForRegion(region, results, false)

	game.battleReceiver <- Battle{Results: calculateTotals(results)}
}

func (game *Game) calculateMultiplayerBattle(region Region, includeDefender bool) {
	results := make(map[PlayerFaction]Result, len(region.IncomingMoves))

	for _, move := range region.IncomingMoves {
		results[move.Faction] = Result{
			Parts: attackModifiers(move, region, true, false, includeDefender),
			Move:  move,
		}
	}

	if !region.isEmpty() && includeDefender {
		results[region.Unit.Faction] = Result{
			Parts:          defenseModifiers(region),
			DefenderRegion: region.Name,
		}
	}

	game.callSupportForRegion(region, results, includeDefender)

	game.battleReceiver <- Battle{Results: calculateTotals(results)}
}

// Battle where units from two regions attack each other simultaneously.
func (game *Game) calculateBorderBattle(region1 Region, region2 Region) {
	move1 := region1.Order
	move2 := region2.Order
	results := map[PlayerFaction]Result{
		move1.Faction: {Parts: attackModifiers(move1, region2, true, true, false), Move: move1},
		move2.Faction: {Parts: attackModifiers(move2, region1, true, true, false), Move: move2},
	}

	// TODO: make these two calls run concurrently
	game.callSupportForRegion(region1, results, false)
	game.callSupportForRegion(region2, results, false)

	game.battleReceiver <- Battle{Results: calculateTotals(results)}
}

type supportDeclaration struct {
	from PlayerFaction
	to   PlayerFaction // Blank if nobody were supported.
}

// Calls support from support orders to the given region, and adds modifiers to the result map.
func (game *Game) callSupportForRegion(
	region Region,
	results map[PlayerFaction]Result,
	includeDefender bool,
) {
	supports := region.IncomingSupports
	supportCount := len(supports)
	supportReceiver := make(chan supportDeclaration, supportCount)

	var waitGroup sync.WaitGroup
	waitGroup.Add(supportCount)

	incomingMoves := []Order{}
	for _, result := range results {
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
			includeDefender,
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

		result, isFaction := results[support.to]
		if isFaction {
			result.Parts = append(
				result.Parts,
				Modifier{Type: ModifierSupport, Value: 1, SupportingFaction: support.from},
			)
			results[support.to] = result
		}
	}
}

func (game *Game) callSupportFromPlayer(
	support Order,
	region Region,
	moves []Order,
	includeDefender bool,
	supportReceiver chan<- supportDeclaration,
	waitGroup *sync.WaitGroup,
) {
	defer waitGroup.Done()

	if includeDefender && !region.isEmpty() && region.Unit.Faction == support.Faction {
		supportReceiver <- supportDeclaration{from: support.Faction, to: support.Faction}
		return
	}

	for _, move := range moves {
		if support.Faction == move.Faction {
			supportReceiver <- supportDeclaration{from: support.Faction, to: support.Faction}
			return
		}
	}

	battlers := make([]PlayerFaction, 0, len(moves)+1)
	for _, move := range moves {
		battlers = append(battlers, move.Faction)
	}
	if includeDefender && !region.isEmpty() {
		battlers = append(battlers, region.Unit.Faction)
	}

	if err := game.messenger.SendSupportRequest(
		support.Faction,
		support.Origin,
		region.Name,
		battlers,
	); err != nil {
		game.log.ErrorCause(err, "failed to send support request")
		supportReceiver <- supportDeclaration{from: support.Faction, to: ""}
		return
	}

	supported, err := game.messenger.AwaitSupport(support.Faction, support.Origin, region.Name)
	if err != nil {
		game.log.ErrorCausef(
			err,
			"failed to receive support declaration from faction '%s'",
			support.Faction,
		)
		supportReceiver <- supportDeclaration{from: support.Faction, to: ""}
		return
	}

	supportReceiver <- supportDeclaration{from: support.Faction, to: supported}
}

func calculateTotals(results map[PlayerFaction]Result) []Result {
	resultList := make([]Result, 0, len(results))

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

func (game *Game) resolveBattle(battle Battle) {
	if battle.isBorderConflict() {
		game.resolveBorderBattle(battle)
	} else if len(battle.Results) == 1 {
		game.resolveSingleplayerBattle(battle)
	} else {
		game.resolveMultiplayerBattle(battle)
	}

	game.resolvedBattles = append(game.resolvedBattles, battle)

	for _, region := range battle.regionNames() {
		game.resolving.Remove(region)
	}
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
			game.Board.removeUnit(move.Unit, move.Origin)
		}
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
}

func (game *Game) resolveMultiplayerBattle(battle Battle) {
	winners, losers := battle.winnersAndLosers()
	tie := len(winners) != 1

	for _, result := range battle.Results {
		// If the result has a DefenderRegion, it is the result of the region's defender.
		// If the defender won or tied, nothing changes for them.
		// If an attacker won, changes to the defender will be handled by calling succeedMove.
		if result.DefenderRegion != "" {
			continue
		}

		move := result.Move

		lost := false
		for _, otherFaction := range losers {
			if otherFaction == move.Faction {
				lost = true
			}
		}

		if lost {
			game.Board.removeOrder(move)
			game.Board.removeUnit(move.Unit, move.Origin)
			continue
		}

		if tie {
			game.Board.removeOrder(move)
			game.attemptRetreat(move)
			continue
		}

		if game.Board[move.Destination].isControlled() {
			game.succeedMove(move)
		}
	}
}

func (battle Battle) isBorderConflict() bool {
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
		game.retreats[move.Origin] = move
		return
	}

	origin.Unit = move.Unit
	game.Board[move.Origin] = origin
}
