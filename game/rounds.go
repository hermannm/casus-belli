package game

import (
	"errors"

	"github.com/immerse-ntnu/hermannia/server/messages"
)

// Initializes a new round of the game.
func (game *Game) NewRound() {
	var season Season
	if len(game.Rounds) == 0 {
		season = Winter
	} else {
		season = nextSeason(game.Rounds[len(game.Rounds)-1].Season)
	}

	round := Round{
		Season:       season,
		FirstOrders:  make([]*Order, 0),
		SecondOrders: make([]*Order, 0),
	}
	game.Rounds = append(game.Rounds, &round)

	// Waits for submitted orders from each player, then adds them to the round.
	received := make(chan []Order, len(game.Messages))
	for player, receiver := range game.Messages {
		timeout := make(chan struct{})
		go game.receiveAndValidateOrders(player, receiver, season, received, timeout)
	}
	for orderSet := range received {
		game.addOrders(orderSet)
	}

	game.Board.Resolve(&round)
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
			parsed, err := parseSubmittedOrders(submitted.Orders, player, game.Board)
			if err != nil {
				game.Lobby.Players[string(player)].Send(err.Error())
				continue
			}

			err = validateOrderSet(parsed, season)
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
func parseSubmittedOrders(submitted []messages.Order, player Player, board Board) ([]Order, error) {
	parsed := make([]Order, 0)
	for _, submittedOrder := range submitted {
		fromArea, ok := board[submittedOrder.From]
		if !ok {
			return nil, errors.New("invalid order origin area")
		}
		toArea, ok := board[submittedOrder.To]
		if !ok {
			return nil, errors.New("invalid order destiantion area")
		}

		parsed = append(parsed, Order{
			Type:   OrderType(submittedOrder.OrderType),
			Player: player,
			From:   fromArea,
			To:     toArea,
			Via:    submittedOrder.Via,
			Build:  UnitType(submittedOrder.Build),
			Status: Pending,
		})
	}

	return parsed, nil
}

// Receives orders to be processed in the current round.
// Sorts orders on their sequence in the round.
func (game *Game) addOrders(orders []Order) {
	round := game.Rounds[len(game.Rounds)-1]

	for _, order := range orders {
		// If order origin has no unit, or unit of different color,
		// then order is a second horse move and should be processed after all others.
		if order.From.IsEmpty() || order.From.Unit.Player != order.Player {
			round.SecondOrders = append(round.SecondOrders, &order)
			continue
		}

		round.FirstOrders = append(round.SecondOrders, &order)
	}
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
