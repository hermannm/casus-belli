package game

type DangerZone string

// Result of an order that had to cross a danger zone to its destination, rolling dice to succeed.
// For move orders, the moved unit dies if it fails the crossing.
// For support orders, the support is cut if it fails the crossing.
type DangerZoneCrossing struct {
	DangerZone DangerZone
	Survived   bool
	DiceResult int
	Order      Order
}

// Finds support orders attempting to cross danger zones to their destinations, and fails them if
// they don't make it across.
func (game *Game) resolveDangerZoneSupports() {
	var crossings []DangerZoneCrossing
	for _, region := range game.board {
		crossings = game.resolveDangerZoneCrossings(region, region.incomingSupports, crossings)
	}

	if len(crossings) != 0 {
		if err := game.messenger.SendDangerZoneCrossings(crossings); err != nil {
			game.log.Error(err)
		}
	}
}

// Finds incoming move orders to the given region that attempt to cross danger zones, and kills them
// if they fail.
func (game *Game) resolveDangerZoneMoves(region *Region) {
	crossings := game.resolveDangerZoneCrossings(region, region.incomingMoves, nil)
	if len(crossings) != 0 {
		if err := game.messenger.SendDangerZoneCrossings(crossings); err != nil {
			game.log.Error(err)
		}
	}

	region.dangerZonesResolved = true
}

func (game *Game) resolveDangerZoneCrossings(
	region *Region,
	incomingMovesOrSupports []Order,
	crossingsToAppend []DangerZoneCrossing,
) []DangerZoneCrossing {
	for _, order := range incomingMovesOrSupports {
		neighbor, adjacent := region.getNeighbor(order.Origin, order.ViaDangerZone)

		// Non-adjacent moves are handled by resolveTransports
		if !adjacent || neighbor.DangerZone == "" {
			continue
		}

		crossing := game.crossDangerZone(order, neighbor.DangerZone)
		crossingsToAppend = append(crossingsToAppend, crossing)

		if !crossing.Survived {
			if order.Type == OrderMove {
				game.board.killMove(order)
			} else {
				game.board.removeOrder(order)
			}
		}
	}

	return crossingsToAppend
}

// Number to beat when attempting to cross a danger zone.
const MinDiceResultToSurviveDangerZone = 3

// Rolls dice to see if order makes it across danger zone.
func (game *Game) crossDangerZone(order Order, dangerZone DangerZone) DangerZoneCrossing {
	diceResult := game.rollDice()
	return DangerZoneCrossing{
		DangerZone: dangerZone,
		Survived:   diceResult >= MinDiceResultToSurviveDangerZone,
		DiceResult: diceResult,
		Order:      order,
	}
}
