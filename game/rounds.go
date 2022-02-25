package game

import (
	"hermannm.dev/bfh-server/game/board"
	"hermannm.dev/bfh-server/game/messages"
	"hermannm.dev/bfh-server/game/validation"
)

// Initializes a new round of the game.
func (game *Game) NewRound() {
	var season board.Season
	if len(game.Rounds) == 0 {
		season = board.Winter
	} else {
		season = nextSeason(game.Rounds[len(game.Rounds)-1].Season)
	}

	// Waits for submitted orders from each player, then adds them to the round.
	received := make(chan []board.Order, len(game.Messages))
	for player, receiver := range game.Messages {
		timeout := make(chan struct{})
		go game.receiveAndValidateOrders(player, receiver, season, received, timeout)
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

	game.Board.Resolve(round)
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
// Function stops if it receives on the timeout channel.
func (game *Game) receiveAndValidateOrders(
	playerID string,
	receiver *messages.Receiver,
	season board.Season,
	output chan<- []board.Order,
	timeout <-chan struct{},
) {
	for {
		select {
		case submitted := <-receiver.Orders:
			parsed, err := parseSubmittedOrders(submitted.Orders, playerID)
			if err != nil {
				game.Lobby.Players[playerID].Send(err.Error())
				continue
			}

			err = validation.ValidateOrderSet(parsed, game.Board, season)
			if err != nil {
				game.Lobby.Players[playerID].Send(err.Error())
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
func parseSubmittedOrders(submitted []messages.Order, playerID string) ([]board.Order, error) {
	parsed := make([]board.Order, 0)

	for _, submittedOrder := range submitted {
		parsed = append(parsed, board.Order{
			Type:   board.OrderType(submittedOrder.OrderType),
			Player: board.Player(playerID),
			From:   submittedOrder.From,
			To:     submittedOrder.To,
			Via:    submittedOrder.Via,
			Build:  board.UnitType(submittedOrder.Build),
		})
	}

	return parsed, nil
}

// Takes a set of orders, and sorts them into two sets based on their sequence in the round.
// Also takes the board for deciding the sequence.
func sortOrders(allOrders []board.Order, brd board.Board) (firstOrders []board.Order, secondOrders []board.Order) {
	firstOrders = make([]board.Order, 0)
	secondOrders = make([]board.Order, 0)

	for _, order := range allOrders {
		fromArea := brd[order.From]

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
	case board.Winter:
		return board.Spring
	case board.Spring:
		return board.Summer
	case board.Summer:
		return board.Fall
	case board.Fall:
		return board.Winter
	default:
		return board.Winter
	}
}
