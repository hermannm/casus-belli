package ordervalidation

import (
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/devlog/log"
)

type Messenger interface {
	AwaitOrders(fromPlayer string) ([]gametypes.Order, error)
	SendOrderRequest(toPlayer string) error
	SendOrdersConfirmation(playerWhoSubmittedOrders string) error
	SendOrdersReceived(playerOrders map[string][]gametypes.Order) error
	SendError(toPlayer string, err error)
}

func GatherAndValidateOrders(
	players []string, board gametypes.Board, season gametypes.Season, messenger Messenger,
) []gametypes.Order {
	orderChans := make(map[string]chan []gametypes.Order)
	for _, player := range players {
		orderChan := make(chan []gametypes.Order, 1)
		orderChans[player] = orderChan
		go gatherAndValidateOrderSet(player, board, season, orderChan, messenger)
	}

	var allOrders []gametypes.Order
	playerOrders := make(map[string][]gametypes.Order)
	for player, orderChan := range orderChans {
		orders := <-orderChan
		allOrders = append(allOrders, orders...)
		playerOrders[player] = orders
	}

	if err := messenger.SendOrdersReceived(playerOrders); err != nil {
		log.Error(err, "")
	}

	return allOrders
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
func gatherAndValidateOrderSet(
	player string,
	board gametypes.Board,
	season gametypes.Season,
	orderChan chan<- []gametypes.Order,
	messenger Messenger,
) {
	for {
		if err := messenger.SendOrderRequest(player); err != nil {
			log.Error(err, "")
			orderChan <- []gametypes.Order{}
			return
		}

		orders, err := messenger.AwaitOrders(player)
		if err != nil {
			log.Error(err, "")
			orderChan <- []gametypes.Order{}
			return
		}

		for i, order := range orders {
			order.Player = player

			origin, ok := board.Regions[order.Origin]
			if ok && !origin.IsEmpty() && order.Type != gametypes.OrderBuild {
				order.Unit = origin.Unit
			}

			orders[i] = order
		}

		if err := validateOrders(orders, board, season); err != nil {
			log.Error(err, "")
			messenger.SendError(player, err)
			continue
		}

		if err := messenger.SendOrdersConfirmation(player); err != nil {
			log.Error(err, "")
		}

		orderChan <- orders
	}
}

// Checks if the given set of orders are valid for the state of the board in the given season.
// Assumes that all orders are from the same player.
func validateOrders(
	orders []gametypes.Order, board gametypes.Board, season gametypes.Season,
) error {
	var err error
	if season == gametypes.SeasonWinter {
		err = validateWinterOrders(orders, board)
	} else {
		err = validateNonWinterOrders(orders, board)
	}
	return err
}
