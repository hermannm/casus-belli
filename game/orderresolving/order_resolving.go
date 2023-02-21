package orderresolving

import (
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
)

type Messenger interface {
	SendBattleResults(battles []gametypes.Battle) error
	SendSupportRequest(to string, supportingRegion string, battlers []string) error
	ReceiveSupport(from string, fromRegion string) (supportTo string, err error)
}

// Adds the round's orders to the board, and resolves them.
// Returns a list of any potential battles from the round.
func ResolveOrders(
	board gametypes.Board, orders []gametypes.Order, season gametypes.Season, messenger Messenger,
) (battles []gametypes.Battle, winner string, hasWinner bool) {
	battles = make([]gametypes.Battle, 0)

	if season == gametypes.SeasonWinter {
		resolveWinterOrders(board, orders)
		return nil, "", false
	}

	firstOrders, secondOrders := SortNonWinterOrders(orders, board)

	firstBattles := resolveNonWinterOrders(board, firstOrders, messenger)
	battles = append(battles, firstBattles...)

	secondBattles := resolveNonWinterOrders(board, secondOrders, messenger)
	battles = append(battles, secondBattles...)

	resolveSieges(board)

	winner, hasWinner = board.CheckWinner()

	return battles, winner, hasWinner
}

// Takes a set of orders, and sorts them into two sets based on their sequence in the round.
// Also takes the board for deciding the sequence.
func SortNonWinterOrders(
	orders []gametypes.Order, board gametypes.Board,
) (firstOrders []gametypes.Order, secondOrders []gametypes.Order) {
	for _, order := range orders {
		fromRegion := board.Regions[order.From]

		// If order origin has no unit, or unit of different color,
		// then order is a second horse move and should be processed after all others.
		if fromRegion.IsEmpty() || fromRegion.Unit.Player != order.Player {
			secondOrders = append(secondOrders, order)
		} else {
			firstOrders = append(firstOrders, order)
		}
	}

	return firstOrders, secondOrders
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
func resolveWinterOrders(board gametypes.Board, orders []gametypes.Order) {
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
