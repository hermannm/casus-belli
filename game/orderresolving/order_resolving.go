package orderresolving

import (
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
)

// A set of player-submitted orders for a round of the game.
type Round struct {
	// Affects the type of orders that can be played in the round.
	Season gametypes.Season `json:"season"`

	// The main set of orders for the round.
	FirstOrders []gametypes.Order `json:"firstOrders"`

	// Set of orders that are known to be executed after the first orders (e.g. horse moves).
	SecondOrders []gametypes.Order `json:"secondOrders"`
}

type Messenger interface {
	SendBattleResults(battles []gametypes.Battle) error
	SendSupportRequest(to string, supportingRegion string, battlers []string) error
	ReceiveSupport(from string, fromRegion string) (supportTo string, err error)
}

// Adds the round's orders to the board, and resolves them.
// Returns a list of any potential battles from the round.
func ResolveOrders(
	board gametypes.Board, round Round, messenger Messenger,
) (battles []gametypes.Battle, winner string, hasWinner bool) {
	battles = make([]gametypes.Battle, 0)

	switch round.Season {
	case gametypes.SeasonWinter:
		resolveWinter(board, round.FirstOrders)
	default:
		firstBattles := resolveNonWinterOrders(board, round.FirstOrders, messenger)
		battles = append(battles, firstBattles...)

		secondBattles := resolveNonWinterOrders(board, round.SecondOrders, messenger)
		battles = append(battles, secondBattles...)

		resolveSieges(board)

		winner, hasWinner = board.CheckWinner()
	}

	board.RemoveOrders()

	return battles, winner, hasWinner
}

// Resolves results of the given orders on the board.
func resolveNonWinterOrders(
	board gametypes.Board, orders []gametypes.Order, messenger Messenger,
) []gametypes.Battle {
	battles := make([]gametypes.Battle, 0)

	board.AddOrders(orders)

	dangerZoneBattles := resolveDangerZones(board)
	battles = append(battles, dangerZoneBattles...)
	err := messenger.SendBattleResults(dangerZoneBattles)
	if err != nil {
		log.Println(err)
	}

	singleplayerBattles := resolveMoves(board, false, messenger)
	battles = append(battles, singleplayerBattles...)

	remainingBattles := resolveMoves(board, true, messenger)
	battles = append(battles, remainingBattles...)

	return battles
}

// Resolves moves on the board. Returns any resulting battles.
// Only resolves battles between players if allowPlayerConflict is true.
func resolveMoves(
	board gametypes.Board, allowPlayerConflict bool, messenger Messenger,
) []gametypes.Battle {
	battles := make([]gametypes.Battle, 0)

	battleReceiver := make(chan gametypes.Battle)
	processing := make(map[string]struct{})
	processed := make(map[string]struct{})
	retreats := make(map[string]gametypes.Order)

OuterLoop:
	for {
		select {
		case battle := <-battleReceiver:
			battles = append(battles, battle)
			messenger.SendBattleResults([]gametypes.Battle{battle})

			newRetreats := resolveBattle(battle, board)
			for _, retreat := range newRetreats {
				retreats[retreat.From] = retreat
			}

			for _, region := range battle.RegionNames() {
				delete(processing, region)
			}
		default:
		BoardLoop:
			for regionName, region := range board.Regions {
				retreat, hasRetreat := retreats[regionName]

				_, isProcessed := processed[regionName]
				if isProcessed && !hasRetreat {
					continue BoardLoop
				}

				_, isProcessing := processing[regionName]
				if isProcessing {
					continue BoardLoop
				}

				for _, move := range region.IncomingMoves {
					transportAttacked, dangerZones := resolveTransports(move, region, board)

					if transportAttacked {
						if allowPlayerConflict {
							continue BoardLoop
						} else {
							processed[regionName] = struct{}{}
						}
					} else if len(dangerZones) > 0 {
						survived, dangerZoneCrossings := crossDangerZones(move, dangerZones)
						if !survived {
							board.RemoveOrder(move)
						}

						battles = append(battles, dangerZoneCrossings...)
						err := messenger.SendBattleResults(dangerZoneCrossings)
						if err != nil {
							log.Println(err)
						}
					}
				}

				moveCount := len(region.IncomingMoves)
				if moveCount == 0 {
					if hasRetreat && region.IsEmpty() {
						region.Unit = retreat.Unit
						board.Regions[regionName] = region
						delete(retreats, regionName)
					}

					processed[region.Name] = struct{}{}
					continue BoardLoop
				}

				resolveRegionMoves(
					region,
					board,
					moveCount,
					allowPlayerConflict,
					battleReceiver,
					processing,
					processed,
					messenger,
				)
			}

			if len(processing) == 0 && len(retreats) == 0 {
				break OuterLoop
			}
		}
	}

	return battles
}

// Finds move and support orders attempting to cross danger zones to their destinations,
// and fail them if they don't make it across.
// Returns a battle result for each danger zone crossing.
func resolveDangerZones(board gametypes.Board) []gametypes.Battle {
	battles := make([]gametypes.Battle, 0)

	for regionName, region := range board.Regions {
		order := region.Order

		if order.Type != gametypes.OrderMove && order.Type != gametypes.OrderSupport {
			continue
		}

		// Checks if the order tries to cross a danger zone.
		destination, adjacent := region.GetNeighbor(order.To, order.Via)
		if !adjacent || destination.DangerZone == "" {
			continue
		}

		// Resolves the danger zone crossing.
		survived, battle := crossDangerZone(order, destination.DangerZone)
		battles = append(battles, battle)

		// If move fails danger zone crossing, the unit dies.
		// If support fails crossing, only the order fails.
		if !survived {
			if order.Type == gametypes.OrderMove {
				region.Unit = gametypes.Unit{}
				board.Regions[regionName] = region
			}

			board.RemoveOrder(order)
		}
	}

	return battles
}

// Goes through regions with siege orders, and updates the region following the siege.
func resolveSieges(board gametypes.Board) {
	for regionName, region := range board.Regions {
		if region.Order.IsNone() || region.Order.Type != gametypes.OrderBesiege {
			continue
		}

		region.SiegeCount++
		if region.SiegeCount == 2 {
			region.ControllingPlayer = region.Unit.Player
			region.SiegeCount = 0
		}

		board.Regions[regionName] = region
	}
}

// Resolves winter orders (builds and internal moves) on the board.
// Assumes they have already been validated.
func resolveWinter(board gametypes.Board, orders []gametypes.Order) {
	for _, order := range orders {
		switch order.Type {
		case gametypes.OrderBuild:
			from := board.Regions[order.From]
			from.Unit = gametypes.Unit{
				Player: order.Player,
				Type:   order.Build,
			}
			board.Regions[order.From] = from
		case gametypes.OrderMove:
			from := board.Regions[order.From]
			to := board.Regions[order.To]

			to.Unit = from.Unit
			from.Unit = gametypes.Unit{}

			board.Regions[order.From] = from
			board.Regions[order.To] = to
		}
	}
}
