package game

import (
	"fmt"
	"log"
	"sync"

	"hermannm.dev/bfh-server/game/board"
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
		players := game.messenger.ReceiverIDs()

		var wg sync.WaitGroup
		wg.Add(len(players))

		received := make(chan receivedOrders, len(players))

		for _, player := range players {
			go game.receiveAndValidateOrders(player, season, received, &wg)
		}

		wg.Wait()

		playerOrders := make(map[string][]board.Order)
		for orderSet := range received {
			playerOrders[orderSet.player] = orderSet.orders
		}
		firstOrders, secondOrders := sortOrders(playerOrders, game.board)

		round := board.Round{
			Season:       season,
			FirstOrders:  firstOrders,
			SecondOrders: secondOrders,
		}

		game.rounds = append(game.rounds, round)

		battles, newWinner := game.board.Resolve(round, game.messenger)
		winner = newWinner

		for _, battle := range battles {
			err := game.messenger.SendBattleResult(battle)
			if err != nil {
				log.Println(err)
			}
		}
	}

	game.messenger.SendWinner(winner)
}

type receivedOrders struct {
	orders []board.Order
	player string
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
func (game Game) receiveAndValidateOrders(
	player string,
	season board.Season,
	output chan<- receivedOrders,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		err := game.messenger.SendOrderRequest(player)
		if err != nil {
			log.Println(err)
			return
		}

		orders, err := game.messenger.ReceiveOrders(player)
		if err != nil {
			log.Println(err)
			return
		}

		for i, order := range orders {
			order.Player = player
			orders[i] = order
		}

		err = validation.ValidateOrderSet(orders, game.board, season)
		if err != nil {
			log.Println(err)
			game.messenger.SendError(player, fmt.Sprintf("Invalid order set: %s", err.Error()))
			continue
		}

		err = game.messenger.SendOrdersConfirmation(player)
		if err != nil {
			log.Println(err)
		}

		output <- receivedOrders{orders: orders, player: player}
	}
}

// Takes a set of orders, and sorts them into two sets based on their sequence in the round.
// Also takes the board for deciding the sequence.
func sortOrders(playerOrders map[string][]board.Order, brd board.Board) (
	firstOrders []board.Order,
	secondOrders []board.Order,
) {
	firstOrders = make([]board.Order, 0)
	secondOrders = make([]board.Order, 0)

	for _, orders := range playerOrders {
		for _, order := range orders {
			fromArea := brd.Areas[order.From]

			// If order origin has no unit, or unit of different color,
			// then order is a second horse move and should be processed after all others.
			if fromArea.IsEmpty() || fromArea.Unit.Player != order.Player {
				secondOrders = append(secondOrders, order)
			} else {
				firstOrders = append(firstOrders, order)
			}
		}
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
