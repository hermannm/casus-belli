package game

// Initializes a new round of the game.
func (game *Game) NewRound() {
	var season Season

	if len(game.Rounds) == 0 {
		season = Winter
	} else {
		season = nextSeason(game.Rounds[len(game.Rounds)-1].Season)
	}

	game.Rounds = append(game.Rounds, &Round{
		Season:       season,
		FirstOrders:  make([]*Order, 0),
		SecondOrders: make([]*Order, 0),
	})
}

// Receives orders to be processed in the current round.
// Sorts orders on their sequence in the round.
func (game *Game) ReceiveOrders(orders []Order) {
	round := game.Rounds[len(game.Rounds)-1]

	round.mut.Lock()
	defer round.mut.Unlock()

	for _, order := range orders {
		// If order origin has no unit, or unit of different color,
		// then order is a second horse move and should be processed after all others.
		if order.From.Unit == nil {
			round.SecondOrders = append(round.SecondOrders, &order)
		} else {
			if order.From.Unit.Color == order.Player {
				round.FirstOrders = append(round.SecondOrders, &order)
			} else {
				round.SecondOrders = append(round.SecondOrders, &order)
			}
		}
	}
}

// Helper function to get the next season given the current season.
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
