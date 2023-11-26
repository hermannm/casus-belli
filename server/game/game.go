package game

import (
	"math/rand"

	"hermannm.dev/devlog/log"
)

type Game struct {
	BoardInfo
	board          Board
	season         Season
	battleReceiver chan Battle
	messenger      Messenger
	log            log.Logger
	rollDice       func() int
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
	SendDangerZoneCrossings(crossings []DangerZoneCrossing) error
	SendWinner(winner PlayerFaction) error
	ClearMessages()
}

func New(
	board Board,
	boardInfo BoardInfo,
	messenger Messenger,
	logger log.Logger,
	customDiceRoller func() int,
) *Game {
	game := Game{
		board:          board,
		BoardInfo:      boardInfo,
		season:         SeasonWinter,
		battleReceiver: make(chan Battle),
		messenger:      messenger,
		log:            logger,
		rollDice:       customDiceRoller,
	}
	if game.rollDice == nil {
		game.rollDice = func() int {
			return rand.Intn(6) + 1
		}
	}

	return &game
}

func (game *Game) Run() {
	if err := game.messenger.SendGameStarted(game.board); err != nil {
		game.log.Error(err)
	}

	for {
		orders := game.gatherAndValidateOrders()

		if game.season == SeasonWinter {
			game.resolveWinterOrders(orders)
		} else {
			game.resolveNonWinterOrders(orders)

			if winner := game.checkWinner(); winner != "" {
				if err := game.messenger.SendWinner(winner); err != nil {
					game.log.Error(err)
				}
				break
			}
		}

		game.nextRound()
	}
}

func (game *Game) nextRound() {
	game.season = game.season.next()

	game.messenger.ClearMessages()
	game.board.resetResolvingState()
	for len(game.battleReceiver) > 0 {
		// Drains the channel - won't block, since there are no other concurrent channel readers, as
		// this function is called from the same goroutine as the one that read
		<-game.battleReceiver
	}
}

func (game *Game) resolveWinterOrders(orders []Order) {
	game.board.placeOrders(orders)

	allResolved := false
	for !allResolved {
		allResolved = true

		for _, region := range game.board {
			switch region.order.Type {
			case OrderBuild:
				region.Unit = Unit{
					Faction: region.order.Faction,
					Type:    region.order.UnitType,
				}
				region.order = Order{}
			case OrderDisband:
				region.removeUnit()
				region.order = Order{}
			}

			if !region.partOfCycle {
				if cycle := game.board.discoverCycle(region.Name, region); cycle != nil {
					game.board.prepareCycleForResolving(cycle)
				}
			}

			if !region.order.isNone() {
				allResolved = false
				continue
			}

			if len(region.incomingMoves) != 0 {
				move := region.incomingMoves[0] // Max 1 incoming move in winter
				region.Unit = move.unit()
				game.board[move.Origin].removeUnit()
				game.board.removeOrder(move)
			}
		}
	}
}

func (game *Game) resolveNonWinterOrders(orders []Order) {
	game.board.placeOrders(orders)
	game.resolveDangerZoneSupports()
	game.resolveMoves()
	game.resolveSieges()
}

func (game *Game) resolveMoves() {
	allRegionsWaiting := false

	for {
		// If all regions are waiting to resolve, then we are either done resolving or waiting for
		// concurrently resolving regions
		if allRegionsWaiting && !game.board.hasUnresolvedRetreats() {
			if !game.board.hasResolvingRegions() {
				break
			}

			// Wait here instead of in select{}, to avoid busy spinning on default case
			battle := <-game.battleReceiver
			game.resolveBattle(battle)
			allRegionsWaiting = false
		}

		select {
		case battle := <-game.battleReceiver:
			game.resolveBattle(battle)
			allRegionsWaiting = false
		default:
			allRegionsWaiting = true
			for _, region := range game.board {
				if waiting := game.resolveRegionMoves(region); !waiting {
					allRegionsWaiting = false
				}
			}
		}
	}
}

func (game *Game) resolveRegionMoves(region *Region) (waiting bool) {
	// Skips the region if it has already been processed
	if region.resolving || (region.resolved && !region.hasUnresolvedRetreat()) {
		return true
	}

	// Resolves incoming moves that require transport
	if !region.transportsResolved {
		if mustWait := game.resolveTransports(region); mustWait {
			return true
		}
	}

	if !region.dangerZonesResolved {
		game.resolveDangerZoneMoves(region)
	}

	// Resolves any unresolved retreat or incoming second horse moves to the region.
	// If the region is not attacked and has no incoming second horse moves, it is fully resolved.
	if !region.attacked() {
		if region.hasUnresolvedRetreat() {
			region.resolveRetreat()
		}
		if region.expectedSecondHorseMoves == 0 {
			region.resolved = true
		} else if region.expectedSecondHorseMoves == len(region.incomingSecondHorseMoves) {
			game.board.placeSecondHorseMoves(region)
		}
		return false
	}

	// Finds out if the region is part of a cycle (moves in a circle)
	if !region.partOfCycle {
		if cycle := game.board.discoverCycle(region.Name, region); cycle != nil {
			if len(cycle) == 2 && cycle[0].order.Faction != cycle[1].order.Faction {
				// If two opposing units move against each other, they battle in the middle
				game.calculateBorderBattle(cycle[0], cycle[1])
			} else {
				game.board.prepareCycleForResolving(cycle)
			}
			return false
		}
	}

	// A single move to an empty region is either an autosuccess, or a singleplayer battle
	if len(region.incomingMoves) == 1 && region.empty() {
		if region.controlled() || region.Sea {
			game.board.succeedMove(region.incomingMoves[0])
			return false
		}

		game.calculateSingleplayerBattle(region)
		return false
	}

	// If the destination region has an outgoing move order, that must be resolved first
	if region.order.Type == OrderMove {
		return true
	}

	if region.resolvingSecondHorseMoves {
		// Checks if supports are cut by other second horse moves
		if mustWait := game.board.cutSupportsAttackedBySecondHorseMoves(region); mustWait {
			return true
		}
	}

	// If the function has not returned yet, then it must be a multiplayer battle
	game.calculateMultiplayerBattle(region)
	return false
}

func (game *Game) resolveSieges() {
	for _, region := range game.board {
		if region.order.Type != OrderBesiege {
			continue
		}

		region.SiegeCount++
		if region.SiegeCount == 2 {
			region.ControllingFaction = region.Unit.Faction
			region.SiegeCount = 0
		}
	}
}

func (game *Game) checkWinner() (winner PlayerFaction) {
	castleCount := make(map[PlayerFaction]int)

	for _, region := range game.board {
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
