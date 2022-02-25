package game

import (
	"hermannm.dev/bfh-server/messages"
)

// Initializes a new round of the game.
func (game *Game) NewRound() {
	var season Season
	if len(game.Rounds) == 0 {
		season = Winter
	} else {
		season = nextSeason(game.Rounds[len(game.Rounds)-1].Season)
	}

	// Waits for submitted orders from each player, then adds them to the round.
	received := make(chan []Order, len(game.Messages))
	for player, receiver := range game.Messages {
		timeout := make(chan struct{})
		go game.receiveAndValidateOrders(player, receiver, season, received, timeout)
	}
	allOrders := make([]Order, 0)
	for orderSet := range received {
		allOrders = append(allOrders, orderSet...)
	}
	firstOrders, secondOrders := sortOrders(allOrders, game.Board)

	round := Round{
		Season:       season,
		FirstOrders:  firstOrders,
		SecondOrders: secondOrders,
	}

	game.Rounds = append(game.Rounds, round)

	game.Board.Resolve(round)
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
// Function stops if it receives on the timeout channel.
func (game *Game) receiveAndValidateOrders(
	player Player,
	receiver *messages.Receiver,
	season Season,
	output chan<- []Order,
	timeout <-chan struct{},
) {
	for {
		select {
		case submitted := <-receiver.Orders:
			parsed, err := parseSubmittedOrders(submitted.Orders, player)
			if err != nil {
				game.Lobby.Players[string(player)].Send(err.Error())
				continue
			}

			err = validateOrderSet(parsed, game.Board, season)
			if err != nil {
				game.Lobby.Players[string(player)].Send(err.Error())
				continue
			}

			output <- parsed
			return
		case <-timeout:
			return
		}
	}
}

// Takes a set of orders in the raw message format, and parses them to the game format.
// Returns the parsed order set, or an error if the parsing failed.
func parseSubmittedOrders(submitted []messages.Order, player Player) ([]Order, error) {
	parsed := make([]Order, 0)

	for _, submittedOrder := range submitted {
		parsed = append(parsed, Order{
			Type:   OrderType(submittedOrder.OrderType),
			Player: player,
			From:   submittedOrder.From,
			To:     submittedOrder.To,
			Via:    submittedOrder.Via,
			Build:  UnitType(submittedOrder.Build),
		})
	}

	return parsed, nil
}

// Takes a set of orders, and sorts them into two sets based on their sequence in the round.
// Also takes the board for deciding the sequence.
func sortOrders(allOrders []Order, board Board) (firstOrders []Order, secondOrders []Order) {
	firstOrders = make([]Order, 0)
	secondOrders = make([]Order, 0)

	for _, order := range allOrders {
		fromArea := board[order.From]

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
func nextSeason(season Season) Season {
	switch season {
	case Winter:
		return Spring
	case Spring:
		return Summer
	case Summer:
		return Fall
	case Fall:
		return Winter
	default:
		return Winter
	}
}
