package game

import (
	"hermannm.dev/devlog/log"
	"hermannm.dev/set"
)

type Game struct {
	Board     Board
	BoardInfo BoardInfo
	Factions  []PlayerFaction
	messenger Messenger
	log       log.Logger

	season             Season
	resolving          set.ArraySet[RegionName]
	resolved           set.ArraySet[RegionName]
	resolvedTransports set.ArraySet[RegionName]
	resolvedBattles    []Battle
	battleReceiver     chan Battle
	retreats           map[RegionName]Order
	secondHorseMoves   []Order
}

type BoardInfo struct {
	ID                 string
	Name               string
	WinningCastleCount int
}

// A faction on the board (e.g. green/red/yellow units) controlled by a player.
type PlayerFaction string

type Messenger interface {
	SendError(to PlayerFaction, err error)
	SendOrderRequest(to PlayerFaction) error
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
	SendBattleResults(battles []Battle) error
	SendWinner(winner PlayerFaction) error
}

func New(
	board Board,
	boardInfo BoardInfo,
	messenger Messenger,
	logger log.Logger,
) *Game {
	return &Game{
		Board:              board,
		BoardInfo:          boardInfo,
		Factions:           board.playerFactions(),
		messenger:          messenger,
		log:                logger,
		season:             SeasonWinter,
		resolving:          set.NewArraySet[RegionName](),
		resolved:           set.NewArraySet[RegionName](),
		resolvedTransports: set.NewArraySet[RegionName](),
		resolvedBattles:    nil,
		battleReceiver:     make(chan Battle),
		retreats:           make(map[RegionName]Order),
		secondHorseMoves:   nil,
	}
}

func (game *Game) Run() {
	for {
		orders := game.GatherAndValidateOrders()

		if game.season == SeasonWinter {
			game.ResolveWinterOrders(orders)
		} else {
			game.ResolveNonWinterOrders(orders)

			if winner := game.CheckWinner(); winner != "" {
				game.messenger.SendWinner(winner)
				break
			}
		}

		game.NextRound()
	}
}

func (game *Game) NextRound() {
	game.season = game.season.Next()

	game.Board.clearOrders()
	game.resolving.Clear()
	game.resolved.Clear()
	game.resolvedTransports.Clear()
	game.resolvedBattles = game.resolvedBattles[:0] // Keeps same capacity
	for len(game.battleReceiver) > 0 {
		// Drains the channel - won't block, since there are no other concurrent channel readers, as
		// this function is called from the same goroutine as the one that read
		<-game.battleReceiver
	}
	game.retreats = make(map[RegionName]Order)
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
			game.Board[order.Origin] = region
		case OrderMove:
			origin := game.Board[order.Origin]
			destination := game.Board[order.Destination]

			destination.Unit = origin.Unit
			origin.Unit = Unit{}

			game.Board[order.Origin] = origin
			game.Board[order.Destination] = destination
		}
	}
}

func (game *Game) ResolveNonWinterOrders(orders []Order) []Battle {
	var battles []Battle

	game.Board.addOrders(orders)

	dangerZoneBattles := resolveDangerZones(game.Board)
	battles = append(battles, dangerZoneBattles...)
	if err := game.messenger.SendBattleResults(dangerZoneBattles); err != nil {
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
OuterLoop:
	for {
		select {
		case battle := <-game.battleReceiver:
			game.resolveBattle(battle)
			game.messenger.SendBattleResults([]Battle{battle})
		default:
			for _, region := range game.Board {
				game.resolveRegionMoves(region)
			}

			if game.resolving.IsEmpty() && len(game.retreats) == 0 {
				break OuterLoop
			}
		}
	}
}

// Immediately resolves region if it does not require battle. If it does require battle, forwards it
// to appropriate battle calculation functions, which send results to game.battleReceiver.
// Skips region if it depends on other moves to resolve first.
func (game *Game) resolveRegionMoves(region Region) {
	retreat, hasRetreat := game.retreats[region.Name]

	// Skips the region if it has already been processed
	if (game.resolved.Contains(region.Name) && !hasRetreat) ||
		(game.resolving.Contains(region.Name)) {
		return
	}

	// Resolves incoming moves that require transport
	if !game.resolvedTransports.Contains(region.Name) {
		game.resolvedTransports.Add(region.Name)

		for _, move := range region.IncomingMoves {
			transportMustWait := game.resolveTransport(move)
			if transportMustWait {
				return
			}
		}
	}

	// Resolves retreats if region has no attackers
	if !region.isAttacked() {
		if hasRetreat && region.isEmpty() {
			region.Unit = retreat.Unit
			game.Board[region.Name] = region
			delete(game.retreats, region.Name)
		}

		game.resolved.Add(region.Name)
		return
	}

	// Finds out if the region is part of a cycle (moves in a circle)
	twoWayCycle, region2, sameFaction := game.Board.discoverTwoWayCycle(region)
	if twoWayCycle && sameFaction {
		// If both moves are by the same player faction, removes the units from their origin
		// regions, as they may not be allowed to retreat if their origin region is taken
		for _, cycleRegion := range [2]Region{region, region2} {
			cycleRegion.Unit = Unit{}
			cycleRegion.Order = Order{}
			game.Board[cycleRegion.Name] = cycleRegion
		}
	} else if twoWayCycle {
		// If the moves are from different player factions, they battle in the middle
		go game.calculateBorderBattle(region, region2)
		game.resolving.AddMultiple(region.Name, region2.Name)
		return
	} else if cycle, _ := game.Board.discoverCycle(region.Name, region.Order); cycle != nil {
		// If there is a cycle longer than 2 moves, forwards the resolving to resolveCycle
		game.resolveCycle(cycle)
		return
	}

	// A single move to an empty region is either an autosuccess, or a singleplayer battle
	if len(region.IncomingMoves) == 1 && region.isEmpty() {
		move := region.IncomingMoves[0]

		if region.isControlled() || region.IsSea {
			game.succeedMove(move)
			return
		}

		go game.calculateSingleplayerBattle(region, move)
		game.resolving.Add(region.Name)
		return
	}

	// If the destination region has an outgoing move order, that must be resolved first
	if region.Order.Type == OrderMove {
		return
	}

	// If the function has not returned yet, then it must be a multiplayer battle
	go game.calculateMultiplayerBattle(region, !region.isEmpty())
	game.resolving.Add(region.Name)
}

func (game *Game) resolveSieges() {
	for regionName, region := range game.Board {
		if region.Order.isNone() || region.Order.Type != OrderBesiege {
			continue
		}

		region.SiegeCount++
		if region.SiegeCount == 2 {
			region.ControllingFaction = region.Unit.Faction
			region.SiegeCount = 0
		}

		game.Board[regionName] = region
	}
}

func (game *Game) succeedMove(move Order) {
	destination := game.Board[move.Destination]

	destination.Unit = move.Unit
	destination.Order = Order{}
	if !destination.IsSea {
		destination.ControllingFaction = move.Faction
	}

	game.Board[move.Destination] = destination

	game.Board.removeUnit(move.Unit, move.Origin)
	game.Board.removeOrder(move)

	game.resolved.Add(move.Destination)

	if secondHorseMove, hasSecondHorseMove := move.tryGetSecondHorseMove(); hasSecondHorseMove {
		game.secondHorseMoves = append(game.secondHorseMoves, secondHorseMove)
	}
}

func (game *Game) addSecondHorseMoves() {
	for _, secondHorseMove := range game.secondHorseMoves {
		game.Board.addOrder(secondHorseMove)
		game.resolved.Remove(secondHorseMove.Destination)
	}

	game.secondHorseMoves = nil
}

func (game *Game) CheckWinner() (winner PlayerFaction) {
	castleCount := make(map[PlayerFaction]int)

	for _, region := range game.Board {
		if region.HasCastle && region.isControlled() {
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

	if tie || highestCount < game.BoardInfo.WinningCastleCount {
		return ""
	}

	return highestCountFaction
}
