package game

import (
	"hermannm.dev/devlog/log"
)

type Game struct {
	BoardInfo
	Board     Board
	messenger Messenger
	log       log.Logger

	season           Season
	resolvedBattles  []Battle
	battleReceiver   chan Battle
	secondHorseMoves []Order
}

type BoardInfo struct {
	ID                 string
	Name               string
	WinningCastleCount int
	PlayerFactions     []PlayerFaction
}

// A faction on the board (e.g. green/red/yellow units) controlled by a player.
type PlayerFaction string

type Messenger interface {
	SendError(to PlayerFaction, err error)
	SendGameStarted(board Board) error
	SendOrderRequest(to PlayerFaction, season Season) error
	AwaitOrders(from PlayerFaction) ([]Order, error)
	SendOrdersReceived(orders map[PlayerFaction][]Order) error
	SendOrdersConfirmation(factionThatSubmittedOrders PlayerFaction) error
	SendSupportRequest(
		to PlayerFaction,
		supporting RegionName,
		embattled RegionName,
		supportable []PlayerFaction,
	) error
	AwaitSupport(
		from PlayerFaction,
		supporting RegionName,
		embattled RegionName,
	) (supported PlayerFaction, err error)
	SendBattleResults(battles ...Battle) error
	SendWinner(winner PlayerFaction) error
	ClearMessages()
}

func New(
	board Board,
	boardInfo BoardInfo,
	messenger Messenger,
	logger log.Logger,
) *Game {
	return &Game{
		Board:            board,
		BoardInfo:        boardInfo,
		messenger:        messenger,
		log:              logger,
		season:           SeasonWinter,
		resolvedBattles:  nil,
		battleReceiver:   make(chan Battle),
		secondHorseMoves: nil,
	}
}

func (game *Game) Run() {
	if err := game.messenger.SendGameStarted(game.Board); err != nil {
		game.log.Error(err)
	}

	for {
		orders := game.gatherAndValidateOrders()

		if game.season == SeasonWinter {
			game.ResolveWinterOrders(orders)
		} else {
			game.ResolveNonWinterOrders(orders)

			if winner := game.checkWinner(); winner != "" {
				game.messenger.SendWinner(winner)
				break
			}
		}

		game.nextRound()
	}
}

func (game *Game) nextRound() {
	game.season = game.season.next()

	game.messenger.ClearMessages()
	game.Board.resetResolvingState()
	game.resolvedBattles = game.resolvedBattles[:0] // Keeps same capacity
	for len(game.battleReceiver) > 0 {
		// Drains the channel - won't block, since there are no other concurrent channel readers, as
		// this function is called from the same goroutine as the one that read
		<-game.battleReceiver
	}
	game.secondHorseMoves = game.secondHorseMoves[:0]
}

func (game *Game) ResolveWinterOrders(orders []Order) {
	for _, order := range orders {
		switch order.Type {
		case OrderBuild:
			region := game.Board[order.Origin]
			region.Unit = Unit{
				Faction: order.Faction,
				Type:    order.Build,
			}
		case OrderMove:
			origin := game.Board[order.Origin]
			destination := game.Board[order.Destination]

			destination.Unit = origin.Unit
			origin.Unit = Unit{}
		}
	}
}

func (game *Game) ResolveNonWinterOrders(orders []Order) []Battle {
	var battles []Battle

	game.Board.addOrders(orders)

	dangerZoneBattles := resolveDangerZones(game.Board)
	battles = append(battles, dangerZoneBattles...)
	if err := game.messenger.SendBattleResults(dangerZoneBattles...); err != nil {
		game.log.Error(err)
	}

	game.resolveMoves()
	game.addSecondHorseMoves()
	game.resolveMoves()
	game.resolveSieges()

	battles = append(battles, game.resolvedBattles...)
	return battles
}

func (game *Game) resolveMoves() {
	anyResolvedInLastLoop := true

	for {
		// If nothing resolved in the last loop and there are no unresolved retreats, then we are
		// either done resolving or waiting for concurrently resolving regions
		if !anyResolvedInLastLoop && !game.Board.hasUnresolvedRetreats() {
			if !game.Board.hasResolvingRegions() {
				break
			}

			// Wait here instead of in select{}, to avoid busy spinning on default case
			battle := <-game.battleReceiver
			game.resolveBattle(battle)
			anyResolvedInLastLoop = true
		}

		select {
		case battle := <-game.battleReceiver:
			game.resolveBattle(battle)
			anyResolvedInLastLoop = true
		default:
			anyResolvedInLastLoop = false
			for _, region := range game.Board {
				if resolved := game.resolveRegionMoves(region); resolved {
					anyResolvedInLastLoop = true
				}
			}
		}
	}
}

func (game *Game) resolveRegionMoves(region *Region) (resolved bool) {
	// Skips the region if it has already been processed
	if region.resolving || (region.resolved && !region.hasRetreat()) {
		return false
	}

	// Resolves incoming moves that require transport
	if !region.transportsResolved {
		region.transportsResolved = true

		for _, move := range region.incomingMoves {
			transportMustWait := game.resolveTransport(move)
			if transportMustWait {
				return false
			}
		}
	}

	// Resolves retreats if region has no attackers
	if !region.attacked() {
		if region.hasRetreat() {
			if region.empty() {
				region.Unit = region.retreat.Unit
			}
			region.retreat = Order{}
		}

		region.resolved = true
		return true
	}

	// Finds out if the region is part of a cycle (moves in a circle)
	twoWayCycle, region2, sameFaction := game.Board.discoverTwoWayCycle(region)
	if twoWayCycle && sameFaction {
		// If both moves are by the same player faction, removes the units from their origin
		// regions, as they may not be allowed to retreat if their origin region is taken
		for _, cycleRegion := range [2]*Region{region, region2} {
			cycleRegion.removeUnit()
			cycleRegion.order = Order{}
			cycleRegion.partOfCycle = true
		}
	} else if twoWayCycle {
		// If the moves are from different player factions, they battle in the middle
		game.calculateBorderBattle(region, region2)
		return true
	} else if cycle, _ := game.Board.discoverCycle(region.Name, region.order); cycle != nil {
		// If there is a cycle longer than 2 moves, forwards the resolving to resolveCycle
		game.resolveCycle(cycle)
		return true
	}

	// A single move to an empty region is either an autosuccess, or a singleplayer battle
	if len(region.incomingMoves) == 1 && region.empty() {
		if region.controlled() || region.Sea {
			game.succeedMove(region.incomingMoves[0])
			return true
		}

		game.calculateSingleplayerBattle(region)
		return true
	}

	// If the destination region has an outgoing move order, that must be resolved first
	if region.order.Type == OrderMove {
		return false
	}

	// If the function has not returned yet, then it must be a multiplayer battle
	game.calculateMultiplayerBattle(region)
	return true
}

func (game *Game) resolveSieges() {
	for _, region := range game.Board {
		if region.order.isNone() || region.order.Type != OrderBesiege {
			continue
		}

		region.SiegeCount++
		if region.SiegeCount == 2 {
			region.ControllingFaction = region.Unit.Faction
			region.SiegeCount = 0
		}
	}
}

func (game *Game) succeedMove(move Order) {
	destination := game.Board[move.Destination]

	destination.replaceUnit(move.Unit)
	destination.order = Order{}
	if !destination.Sea {
		destination.ControllingFaction = move.Faction
	}

	game.Board[move.Origin].removeUnit()
	game.Board.removeOrder(move)

	destination.resolved = true

	if move.hasSecondHorseMove() {
		game.secondHorseMoves = append(game.secondHorseMoves, move.secondHorseMove())
	}
}

func (game *Game) addSecondHorseMoves() {
	for _, secondHorseMove := range game.secondHorseMoves {
		game.Board.addOrder(secondHorseMove)
		destination := game.Board[secondHorseMove.Destination]
		destination.resolved = false
		destination.partOfCycle = false
		destination.transportsResolved = false
	}

	game.secondHorseMoves = nil
}

func (game *Game) checkWinner() (winner PlayerFaction) {
	castleCount := make(map[PlayerFaction]int)

	for _, region := range game.Board {
		if region.Castle && region.controlled() {
			castleCount[region.ControllingFaction]++
		}
	}

	tie := false
	highestCount := 0
	var highestCountFaction PlayerFaction
	for faction, count := range castleCount {
		if count > highestCount {
			highestCount = count
			highestCountFaction = faction
			tie = false
		} else if count == highestCount {
			tie = true
		}
	}

	if tie || highestCount < game.WinningCastleCount {
		return ""
	}

	return highestCountFaction
}
