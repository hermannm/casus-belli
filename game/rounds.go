package game

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

func (game *Game) ReceiveOrders(orders []Order) {
	round := game.Rounds[len(game.Rounds)-1]

	round.mut.Lock()
	defer round.mut.Unlock()

	for _, order := range orders {
		if order.From.Unit == nil {
			round.SecondOrders = append(round.SecondOrders, &order)
		} else {
			if order.From.Unit.Color == order.Player.Color {
				round.FirstOrders = append(round.SecondOrders, &order)
			} else {
				round.SecondOrders = append(round.SecondOrders, &order)
			}
		}
	}
}

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
