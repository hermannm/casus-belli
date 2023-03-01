package game

import (
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/ordervalidation"
)

func (game *Game) gatherAndValidateOrderSets(season gametypes.Season) []gametypes.Order {
	// Waits for submitted orders from each player, then adds them to the round.
	players := game.messenger.ReceiverIDs()

	orderChans := make(map[string]chan []gametypes.Order)
	for _, player := range players {
		orderChan := make(chan []gametypes.Order, 1)
		orderChans[player] = orderChan
		go game.gatherAndValidateOrderSet(player, season, orderChan)
	}

	var allOrders []gametypes.Order
	playerOrders := make(map[string][]gametypes.Order)
	for player, orderChan := range orderChans {
		orders := <-orderChan
		allOrders = append(allOrders, orders...)
		playerOrders[player] = orders
	}

	if err := game.messenger.SendOrdersReceived(playerOrders); err != nil {
		log.Println(err)
	}

	return allOrders
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
func (game Game) gatherAndValidateOrderSet(
	player string, season gametypes.Season, orderChan chan<- []gametypes.Order,
) {
	for {
		err := game.messenger.SendOrderRequest(player)
		if err != nil {
			log.Println(err)
			orderChan <- []gametypes.Order{}
			return
		}

		orders, err := game.messenger.ReceiveOrders(player)
		if err != nil {
			log.Println(err)
			orderChan <- []gametypes.Order{}
			return
		}

		for i, order := range orders {
			order.Player = player
			orders[i] = order
		}

		err = ordervalidation.ValidateOrders(orders, game.board, season)
		if err != nil {
			log.Println(err)
			game.messenger.SendError(player, err.Error())
			continue
		}

		err = game.messenger.SendOrdersConfirmation(player)
		if err != nil {
			log.Println(err)
		}

		orderChan <- orders
	}
}
