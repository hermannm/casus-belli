package game

type DangerZone string

// Number to beat when attempting to cross a danger zone.
const MinResultToSurviveDangerZone = 3

func newDangerZoneCrossing(order Order, dangerZone DangerZone) Battle {
	return Battle{Results: []Result{{Order: order}}, DangerZone: dangerZone}
}

func (game *Game) resolveDangerZoneCrossings(region *Region) {
	if region.dangerZonesResolved {
		return
	}

	for _, orders := range [...][]Order{region.incomingMoves, region.incomingSupports} {
		for _, order := range orders {
			if mustCross, dangerZone := order.mustCrossDangerZone(region); mustCross {
				game.resolveDangerZoneCrossing(newDangerZoneCrossing(order, dangerZone))
			}
		}
	}

	region.dangerZonesResolved = true
}

func (game *Game) resolveDangerZoneCrossing(crossing Battle) {
	order := crossing.Results[0].Order

	game.messenger.SendBattleAnnouncement(crossing)

	ctx, cleanup := newPlayerInputContext()
	defer cleanup()

	if err := game.messenger.AwaitDiceRoll(ctx, order.Faction); err != nil {
		game.log.Error(err)
	}

	crossing.addModifier(order.Faction, Modifier{Type: ModifierDice, Value: game.rollDice()})

	if crossing.Results[0].Total < MinResultToSurviveDangerZone {
		if order.Type == OrderMove {
			game.board.killMove(order)
		} else {
			game.board.removeOrder(order)
		}
	}

	game.messenger.SendBattleResults(crossing)
}
