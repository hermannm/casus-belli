package orderresolving

import (
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
)

type Messenger interface {
	SendBattleResults(battles []gametypes.Battle) error

	SendSupportRequest(
		toPlayer string,
		supportingRegion string,
		embattledRegion string,
		supportablePlayers []string,
	) error

	ReceiveSupport(
		fromPlayer string, supportingRegion string, embattledRegion string,
	) (supportedPlayer string, err error)
}

func ResolveOrders(
	board gametypes.Board, orders []gametypes.Order, season gametypes.Season, messenger Messenger,
) (battles []gametypes.Battle, winner string, hasWinner bool) {
	if season == gametypes.SeasonWinter {
		resolveWinterOrders(board, orders)
		return nil, "", false
	} else {
		battles = resolveNonWinterOrders(board, orders, messenger)
		winner, hasWinner = board.CheckWinner()
		return battles, winner, hasWinner
	}
}

func resolveWinterOrders(board gametypes.Board, orders []gametypes.Order) {
	for _, order := range orders {
		switch order.Type {
		case gametypes.OrderBuild:
			region := board.Regions[order.Origin]
			region.Unit = gametypes.Unit{
				Player: order.Player,
				Type:   order.Build,
			}
			board.Regions[order.Origin] = region
		case gametypes.OrderMove:
			origin := board.Regions[order.Origin]
			destination := board.Regions[order.Destination]

			destination.Unit = origin.Unit
			origin.Unit = gametypes.Unit{}

			board.Regions[order.Origin] = origin
			board.Regions[order.Destination] = destination
		}
	}
}

func resolveNonWinterOrders(
	board gametypes.Board, orders []gametypes.Order, messenger Messenger,
) []gametypes.Battle {
	var battles []gametypes.Battle

	board.AddOrders(orders)

	dangerZoneBattles := resolveDangerZones(board)
	battles = append(battles, dangerZoneBattles...)
	if err := messenger.SendBattleResults(dangerZoneBattles); err != nil {
		log.Println(err)
	}

	resolver := newMoveResolver()
	resolver.resolveMoves(board, messenger)
	resolver.addSecondHorseMoves(board)
	resolver.resolveMoves(board, messenger)

	battles = append(battles, resolver.resolvedBattles...)

	resolveSieges(board)

	board.RemoveOrders()

	return battles
}

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
