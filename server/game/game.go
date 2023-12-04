package game

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"hermannm.dev/devlog/log"
)

type Game struct {
	BoardInfo
	board     Board
	season    Season
	messenger Messenger
	log       log.Logger
	rollDice  func() int
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
	SendGameStarted(board Board)
	SendOrderRequest(to PlayerFaction, season Season) (succeeded bool)
	SendOrdersConfirmation(factionThatSubmittedOrders PlayerFaction)
	SendOrdersReceived(orders map[PlayerFaction][]Order)
	SendBattleAnnouncement(battle Battle)
	SendBattleResults(battle Battle)
	SendWinner(winner PlayerFaction)
	AwaitOrders(ctx context.Context, from PlayerFaction) ([]Order, error)
	AwaitDiceRoll(ctx context.Context, from PlayerFaction) error
	AwaitSupport(
		ctx context.Context,
		from PlayerFaction,
		embattledRegion RegionName,
	) (supported PlayerFaction, err error)
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
		board:     board,
		BoardInfo: boardInfo,
		season:    SeasonWinter,
		messenger: messenger,
		log:       logger,
		rollDice:  customDiceRoller,
	}
	if game.rollDice == nil {
		game.rollDice = func() int {
			return rand.Intn(6) + 1
		}
	}

	return &game
}

func (game *Game) Run() {
	game.messenger.SendGameStarted(game.board)

	for {
		orders := game.gatherAndValidateOrders()

		if game.season == SeasonWinter {
			game.resolveWinterOrders(orders)
		} else {
			game.resolveNonWinterOrders(orders)

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
	game.board.resetResolvingState()
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
				if cycle := game.board.findCycle(region.Name, region); cycle != nil {
					cycle.prepareForResolving()
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

	game.resolveUncontestedRegions()
	for {
		if game.board.resolved() {
			break
		}
		game.resolveContestedRegions()
		game.resolveUncontestedRegions()
	}

	game.resolveSieges()
}

func (game *Game) resolveContestedRegions() {
	for _, region := range game.board {
		if waiting := game.resolveContestedRegion(region); !waiting {
			return
		}
	}
}

func (game *Game) resolveUncontestedRegions() {
	allRegionsWaiting := false
	for !allRegionsWaiting {
		allRegionsWaiting = true

		for _, region := range game.board {
			if waiting := game.resolveUncontestedRegion(region); !waiting {
				allRegionsWaiting = false
			}
		}
	}
}

func (game *Game) resolveUncontestedRegion(region *Region) (waiting bool) {
	if mustWait := game.board.resolveUncontestedTransports(region); mustWait {
		return true
	}

	if !region.attacked() {
		region.resolveRetreat()

		if region.expectedKnightMoves == 0 {
			region.resolved = true
			return true
		} else if region.expectedKnightMoves == len(region.incomingKnightMoves) {
			game.board.placeKnightMoves(region)
			return false
		}
	}

	if borderBattle, _ := game.board.findBorderBattle(region); borderBattle {
		return true
	}

	// Finds out if the region is part of a cycle (moves in a circle)
	if !region.partOfCycle {
		if cycle := game.board.findCycle(region.Name, region); cycle != nil {
			cycle.prepareForResolving()
			return false
		}
	}

	if len(region.incomingMoves) == 1 && region.empty() && (region.controlled() || region.Sea) {
		move := region.incomingMoves[0]
		if mustCross, _ := move.mustCrossDangerZone(region); mustCross {
			return true
		} else {
			game.board.succeedMove(move)
			return false
		}
	}

	return true
}

func (game *Game) resolveContestedRegion(region *Region) (waiting bool) {
	if region.resolved {
		return true
	}
	if mustWait := game.resolveContestedTransports(region); mustWait {
		return true
	}
	game.resolveDangerZoneCrossings(region)

	if borderBattle, secondRegion := game.board.findBorderBattle(region); borderBattle {
		game.resolveBorderBattle(region, secondRegion)
		return false
	}

	// A single move to an empty region is either an autosuccess, or a singleplayer battle
	if len(region.incomingMoves) == 1 && region.empty() {
		if region.controlled() || region.Sea {
			game.board.succeedMove(region.incomingMoves[0])
		} else {
			game.resolveSingleplayerBattle(region)
		}
		return false
	}

	// If the destination region has an outgoing move order, that must be resolved first
	if region.order.Type == OrderMove {
		return true
	}

	if region.resolvingKnightMoves {
		// Checks if supports are cut by other knight moves
		if mustWait := game.board.cutSupportsAttackedByKnightMoves(region); mustWait {
			return true
		}
	}

	// If the function has not returned yet, then it must be a multiplayer battle
	game.resolveMultiplayerBattle(region)
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

func newPlayerInputContext() (ctx context.Context, cleanup context.CancelFunc) {
	return context.WithTimeoutCause(
		context.Background(),
		1*time.Minute,
		errors.New("timed out after 1 minute"),
	)
}
