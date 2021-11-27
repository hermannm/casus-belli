package game

func (game *Game) NewRound() {
	var season Season

	if len(game.Rounds) == 0 {
		season = Winter
	} else {
		season = NextSeason(game.Rounds[len(game.Rounds)-1].Season)
	}

	game.Rounds = append(game.Rounds, &Round{
		Season:       season,
		FirstOrders:  make([]*Order, 0),
		SecondOrders: make([]*Order, 0),
	})
}

func NextSeason(season Season) Season {
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
