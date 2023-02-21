package game

import (
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/bfh-server/game/orderresolving"
	"hermannm.dev/bfh-server/game/validation"
)

// Initializes a new round of the game.
func (game *Game) Start() {
	var season gametypes.Season

	// Starts new rounds until there is a winner.
	for {
		season = season.Next()

		// Waits for submitted orders from each player, then adds them to the round.
		players := game.messenger.ReceiverIDs()

		orderChans := make(map[string]chan []gametypes.Order)
		for _, player := range players {
			orderChan := make(chan []gametypes.Order, 1)
			orderChans[player] = orderChan
			go game.receiveAndValidateOrders(player, season, orderChan)
		}

		playerOrders := make(map[string][]gametypes.Order)
		for player, orderChan := range orderChans {
			orders := <-orderChan
			playerOrders[player] = orders
		}

		err := game.messenger.SendOrdersReceived(playerOrders)
		if err != nil {
			log.Println(err)
		}

		firstOrders, secondOrders := sortOrders(playerOrders, game.board)

		round := orderresolving.Round{
			Season: season, FirstOrders: firstOrders, SecondOrders: secondOrders,
		}

		game.rounds = append(game.rounds, round)

		battles, winner, hasWinner := orderresolving.ResolveOrders(
			game.board, round, game.messenger,
		)

		err = game.messenger.SendBattleResults(battles)
		if err != nil {
			log.Println(err)
		}

		if hasWinner {
			game.messenger.SendWinner(winner)
			break
		}
	}
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
func (game Game) receiveAndValidateOrders(
	player string,
	season gametypes.Season,
	orderChan chan<- []gametypes.Order,
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

		err = validation.ValidateOrders(orders, game.board, season)
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

// Takes a set of orders, and sorts them into two sets based on their sequence in the round.
// Also takes the board for deciding the sequence.
func sortOrders(playerOrders map[string][]gametypes.Order, board gametypes.Board) (
	firstOrders []gametypes.Order,
	secondOrders []gametypes.Order,
) {
	firstOrders = make([]gametypes.Order, 0)
	secondOrders = make([]gametypes.Order, 0)

	for _, orders := range playerOrders {
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
	}

	return firstOrders, secondOrders
}
