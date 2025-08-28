package game

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"hermannm.dev/opt"
	"hermannm.dev/set"
	"hermannm.dev/wrap"
)

// Results of a battle between players, an attempt to conquer a neutral region, or an attempt to
// cross a danger zone.
type Battle struct {
	// If length is one, the battle was a neutral region conquest attempt or danger zone crossing.
	// If length is more than one, the battle was between players.
	Results []Result

	// If the battle is a danger zone crossing: name of the crossed danger zone.
	DangerZone DangerZone
}

// Dice and modifier result for a battle.
type Result struct {
	Total int
	Parts []Modifier

	// If result of a move order to the battle: the move order in question, otherwise empty.
	// If the result is part of a danger zone crossing, the order is either a move or support order.
	Order opt.Option[Order]

	// If result of a defending unit in a region: the faction of the defender, otherwise blank.
	DefenderFaction PlayerFaction `json:",omitempty"`
}

func (battle *Battle) addModifier(faction PlayerFaction, modifier Modifier) {
	for i, result := range battle.Results {
		if result.DefenderFaction == faction ||
			(result.Order.HasValue() && result.Order.Value.Faction == faction) {
			result.Parts = append(result.Parts, modifier)
			result.Total += modifier.Value
			battle.Results[i] = result
			return
		}
	}
}

func (battle Battle) factions() []PlayerFaction {
	factions := make([]PlayerFaction, len(battle.Results))

	for i, result := range battle.Results {
		if result.DefenderFaction != "" {
			factions[i] = result.DefenderFaction
		} else {
			factions[i] = result.Order.Value.Faction
		}
	}

	return factions
}

func (faction PlayerFaction) isFighting(battle *Battle) bool {
	for _, result := range battle.Results {
		if result.DefenderFaction == faction || result.Order.Value.Faction == faction {
			return true
		}
	}
	return false
}

func (game *Game) resolveSingleplayerBattle(region *Region) {
	move := region.incomingMoves[0]
	battle := Battle{Results: []Result{game.newAttackerResult(move, region, true, false)}}

	game.calculateBattle(&battle, region)

	winners, _ := battle.winnersAndLosers()
	if len(winners) == 1 {
		game.board.succeedMove(move)
	} else {
		game.board.retreatMove(move)
	}

	game.messenger.SendBattleResults(battle)
}

func (game *Game) resolveMultiplayerBattle(region *Region) {
	var battle Battle
	for _, move := range region.incomingMoves {
		battle.Results = append(battle.Results, game.newAttackerResult(move, region, false, false))
	}
	if !region.empty() {
		battle.Results = append(battle.Results, game.newDefenderResult(region.Unit.Value))
	}

	game.calculateBattle(&battle, region)

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
				if tie || !region.controlled() {
					region.removeUnit()
				}
			}
			continue
		}

		move := result.Order.Value
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

	game.messenger.SendBattleResults(battle)
}

func (game *Game) calculateBattle(battle *Battle, region *Region) {
	remainingSupports := battle.addAutomaticSupports(region, region.incomingMoves, false)

	game.messenger.SendBattleAnnouncement(*battle)

	ctx, cleanup := newPlayerInputContext()
	defer cleanup()

	// If we have no supports to call, and only 1 combatant, then we can avoid concurrency
	if len(remainingSupports) == 0 && len(battle.Results) == 1 {
		faction := battle.Results[0].Order.Value.Faction // If 1 result, it must be a move order
		if err := game.messenger.AwaitDiceRoll(ctx, faction); err != nil {
			game.handleBattleError(wrap.Error(err, "failed to receive dice roll"), faction, battle)
		}
		battle.addModifier(faction, Modifier{Type: ModifierDice, Value: game.rollDice()})
	} else {
		var resultsLock sync.Mutex
		var waitGroup sync.WaitGroup

		factionsInBattle := battle.factions()

		for _, faction := range game.PlayerFactions {
			if faction.isFighting(battle) {
				waitGroup.Add(1)
				go game.awaitDiceRoll(ctx, faction, battle, &waitGroup, &resultsLock)
				continue
			}

			supportCount := countOrdersFromFaction(remainingSupports, faction)
			if supportCount != 0 {
				waitGroup.Add(1)
				go game.awaitSupport(
					ctx,
					faction,
					battle,
					region.Name,
					supportCount,
					factionsInBattle,
					&waitGroup,
					&resultsLock,
				)
			}
		}

		waitGroup.Wait()
	}
}

// Battle where units from two regions attack each other simultaneously.
func (game *Game) resolveBorderBattle(region1 *Region, region2 *Region) {
	moveToRegion1, moveToRegion2 := region2.order.Value, region1.order.Value
	battle := Battle{
		Results: []Result{
			game.newAttackerResult(moveToRegion1, region1, false, true),
			game.newAttackerResult(moveToRegion2, region2, false, true),
		},
	}

	game.calculateBorderBattle(&battle, region1, region2)

	winners, losers := battle.winnersAndLosers()

	// If battle was a tie, both moves retreat
	if len(winners) > 1 {
		// Remove both orders before retreating, so they don't think their origins are attacked
		game.board.removeOrder(region1.order.Value)
		game.board.removeOrder(region2.order.Value)

		game.board.retreatMove(region1.order.Value)
		game.board.retreatMove(region2.order.Value)
	} else {
		for _, result := range battle.Results {
			// Only the loser is affected by the results of the border battle; the winner may still
			// have to win a battle in the destination region, which will be handled by the next
			// cycle of move resolving.
			if result.Order.Value.Faction == losers[0] {
				game.board.killMove(result.Order.Value)
				break
			}
		}
	}

	game.messenger.SendBattleResults(battle)
}

var errSupportedOtherRegion error = errors.New("supported other region in border battle")

func (game *Game) calculateBorderBattle(battle *Battle, region1 *Region, region2 *Region) {
	remainingSupports1 := battle.addAutomaticSupports(region1, []Order{region2.order.Value}, true)
	remainingSupports2 := battle.addAutomaticSupports(region2, []Order{region1.order.Value}, true)

	game.messenger.SendBattleAnnouncement(*battle)

	ctx, cleanup := newPlayerInputContext()
	defer cleanup()

	var resultsLock sync.Mutex
	var waitGroup sync.WaitGroup

	for _, faction := range game.PlayerFactions {
		if faction.isFighting(battle) {
			waitGroup.Add(1)
			go game.awaitDiceRoll(ctx, faction, battle, &waitGroup, &resultsLock)
			continue
		}

		supportCount1 := countOrdersFromFaction(remainingSupports1, faction)
		supportCount2 := countOrdersFromFaction(remainingSupports2, faction)

		ctx, cancel := context.WithCancelCause(ctx)

		if supportCount1 != 0 {
			waitGroup.Add(1)
			go func() {
				game.awaitSupport(
					ctx,
					faction,
					battle,
					region1.Name,
					supportCount1,
					[]PlayerFaction{region2.order.Value.Faction},
					&waitGroup,
					&resultsLock,
				)
				cancel(errSupportedOtherRegion)
			}()
		}

		if supportCount2 != 0 {
			waitGroup.Add(1)
			go func() {
				game.awaitSupport(
					ctx,
					faction,
					battle,
					region2.Name,
					supportCount2,
					[]PlayerFaction{region1.order.Value.Faction},
					&waitGroup,
					&resultsLock,
				)
				cancel(errSupportedOtherRegion)
			}()
		}
	}

	waitGroup.Wait()
}

func (game *Game) awaitDiceRoll(
	ctx context.Context,
	faction PlayerFaction,
	battle *Battle,
	waitGroup *sync.WaitGroup,
	resultsLock *sync.Mutex,
) {
	defer waitGroup.Done()

	if err := game.messenger.AwaitDiceRoll(ctx, faction); err != nil {
		game.handleBattleError(wrap.Error(err, "failed to receive dice roll"), faction, battle)
		return
	}

	resultsLock.Lock()
	battle.addModifier(faction, Modifier{Type: ModifierDice, Value: game.rollDice()})
	resultsLock.Unlock()
}

func (game *Game) awaitSupport(
	ctx context.Context,
	faction PlayerFaction,
	battle *Battle,
	regionName RegionName,
	supportCount int,
	supportableFactions []PlayerFaction,
	waitGroup *sync.WaitGroup,
	resultsLock *sync.Mutex,
) {
	defer waitGroup.Done()

	supported, err := game.messenger.AwaitSupport(ctx, faction, regionName)
	if err != nil {
		if err != errSupportedOtherRegion {
			game.handleBattleError(
				wrap.Error(err, "failed to receive support declaration"),
				faction,
				battle,
			)
		}
		return
	}

	if supported == "" {
		return
	}

	if !slices.Contains(supportableFactions, supported) {
		game.handleBattleError(
			fmt.Errorf("received invalid supported faction '%s'", supported),
			faction,
			battle,
		)
		return
	}

	resultsLock.Lock()
	battle.addModifier(
		supported, Modifier{
			Type:              ModifierSupport,
			Value:             supportCount,
			SupportingFaction: faction,
		},
	)
	resultsLock.Unlock()
}

func (game *Game) handleBattleError(err error, faction PlayerFaction, battle *Battle) {
	game.messenger.SendError(faction, err)
	game.log.WarnError(nil, err, "", "from", faction, "battle", battle.regionNames())
}

// Adds modifiers for support orders from players involved in the battle, as we assume they always
// want to support themselves, so we don't need to ask. Returns support orders that could not be
// added automatically.
func (battle *Battle) addAutomaticSupports(
	region *Region,
	incomingMoves []Order,
	borderBattle bool,
) (remainingSupports []Order) {
	supportCounts := make(map[PlayerFaction]int)

SupportLoop:
	for _, support := range region.incomingSupports {
		if !region.empty() && !borderBattle && support.Faction == region.Unit.Value.Faction {
			supportCounts[support.Faction]++
			continue
		}

		for _, move := range incomingMoves {
			if move.Faction == support.Faction {
				supportCounts[support.Faction]++
				continue SupportLoop
			}
		}

		remainingSupports = append(remainingSupports, support)
	}

	for faction, supportCount := range supportCounts {
		battle.addModifier(
			faction,
			Modifier{Type: ModifierSupport, Value: supportCount, SupportingFaction: faction},
		)
	}

	return remainingSupports
}

// Number to beat when attempting to conquer a neutral region.
const MinResultToConquerNeutralRegion int = 4

// In case of a battle against an unconquered region or a danger zone, only one player faction is
// returned in one of the lists.
//
// In case of a battle between players, multiple winners are returned in the case of a tie.
func (battle Battle) winnersAndLosers() (winners []PlayerFaction, losers []PlayerFaction) {
	if len(battle.Results) == 1 {
		result := battle.Results[0]

		if result.Total >= MinResultToConquerNeutralRegion {
			return []PlayerFaction{result.Order.Value.Faction}, nil
		} else {
			return nil, []PlayerFaction{result.Order.Value.Faction}
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
			faction = result.Order.Value.Faction
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
		if result.Order.HasValue() {
			nameSet.Add(result.Order.Value.Destination)
		}
	}

	return nameSet.ToSlice()
}
