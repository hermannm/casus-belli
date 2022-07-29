package game

import (
	"fmt"
	"log"

	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/game/messages"
	"hermannm.dev/bfh-server/game/validation"
)

// Initializes a new round of the game.
func (game *Game) Start() {
	var season board.Season
	var winner string

	// Starts new rounds until there is a winner.
	for winner == "" {
		season = nextSeason(season)

		// Waits for submitted orders from each player, then adds them to the round.
		players := game.msgHandler.ReceiverIDs()
		received := make(chan []board.Order, len(players))
		for _, player := range players {
			go game.receiveAndValidateOrders(player, season, received, game.msgHandler)
		}
		allOrders := make([]board.Order, 0)
		for orderSet := range received {
			allOrders = append(allOrders, orderSet...)
		}
		firstOrders, secondOrders := sortOrders(allOrders, game.Board)

		round := board.Round{
			Season:       season,
			FirstOrders:  firstOrders,
			SecondOrders: secondOrders,
		}

		game.Rounds = append(game.Rounds, round)

		battles, newWinner := game.Board.Resolve(round, game.msgHandler)
		winner = newWinner

		for _, battle := range battles {
			err := game.msgHandler.SendBattleResult(battle)
			if err != nil {
				log.Println(err)
			}
		}
	}

	game.msgHandler.SendWinner(winner)
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
// Function stops if it receives on the timeout channel.
func (game Game) receiveAndValidateOrders(
	player string,
	season board.Season,
	output chan<- []board.Order,
	msgHandler messages.Handler,
) {
	for {
		err := msgHandler.SendOrderRequest(player)
		if err != nil {
			log.Println(err)
			output <- make([]board.Order, 0)
			return
		}

		orders, err := msgHandler.ReceiveOrders(player)
		if err != nil {
			log.Println(err)
			output <- make([]board.Order, 0)
			return
		}

		addOrderPlayer(orders, player)

		err = validation.ValidateOrderSet(orders, game.Board, season)
		if err != nil {
			log.Println(err)
			msgHandler.SendError(player, fmt.Sprintf("Invalid order set: %s", err.Error()))
			continue
		}

		output <- orders
	}
}

// Takes a set of orders, and mutates it by adding the given player ID
// to the Player field on every order.
func addOrderPlayer(orders []board.Order, playerID string) {
	for i, order := range orders {
		order.Player = playerID
		orders[i] = order
	}
}

// Takes a set of orders, and sorts them into two sets based on their sequence in the round.
// Also takes the board for deciding the sequence.
func sortOrders(allOrders []board.Order, brd board.Board) (firstOrders []board.Order, secondOrders []board.Order) {
	firstOrders = make([]board.Order, 0)
	secondOrders = make([]board.Order, 0)

	for _, order := range allOrders {
		fromArea := brd.Areas[order.From]

		// If order origin has no unit, or unit of different color,
		// then order is a second horse move and should be processed after all others.
		if fromArea.IsEmpty() || fromArea.Unit.Player != order.Player {
			secondOrders = append(secondOrders, order)
			continue
		}

		firstOrders = append(firstOrders, order)
	}

	return firstOrders, secondOrders
}

// Returns the next season given the current season.
func nextSeason(season board.Season) board.Season {
	switch season {
	case board.SeasonWinter:
		return board.SeasonSpring
	case board.SeasonSpring:
		return board.SeasonSummer
	case board.SeasonSummer:
		return board.SeasonFall
	case board.SeasonFall:
		return board.SeasonWinter
	default:
		return board.SeasonWinter
	}
}
