package board

// Checks whether the order is initialized.
func (order Order) IsNone() bool {
	return order.Type == ""
}

// Moves the unit of the given move order to its destination,
// killing any unit that may have already been there,
// and sets control of the region to the order's player.
//
// Then removes references to this move on the board,
// and removes any potential order from the destination region.
func (board Board) succeedMove(move Order) {
	to := board.Regions[move.To]

	to.Unit = move.Unit
	to.Order = Order{}
	if !to.Sea {
		to.ControllingPlayer = move.Player
	}
	board.Regions[move.To] = to

	board.removeOriginUnit(move)

	board.removeMove(move)
}

// Removes the move order's unit from its origin region, if it still exists.
func (board Board) removeOriginUnit(move Order) {
	from := board.Regions[move.From]

	if move.Unit == from.Unit {
		from.Unit = Unit{}
		board.Regions[move.From] = from
	}
}

// Rolls dice to see if order makes it across danger zone.
// Returns whether the order succeeded, and the resulting battle for use by the client.
func (order Order) crossDangerZone(dangerZone string) (survived bool, result Battle) {
	diceMod := Modifier{Type: ModifierDice, Value: rollDice()}

	// Records crossing attempt as a battle, so clients can see dice roll.
	battle := Battle{
		Results:    []Result{{Total: diceMod.Value, Parts: []Modifier{diceMod}, Move: order}},
		DangerZone: dangerZone,
	}

	// All danger zones currently require a dice roll greater than 2.
	// May need to be changed in the future if a more dynamic implementation is preferred.
	return diceMod.Value > 2, battle
}

func (order Order) crossDangerZones(dangerZones []string) (survivedAll bool, results []Battle) {
	survivedAll = true
	results = make([]Battle, 0)

	for _, dangerZone := range dangerZones {
		survived, result := order.crossDangerZone(dangerZone)
		results = append(results, result)
		if !survived {
			survivedAll = false
		}
	}

	return survivedAll, results
}

func (board Board) addOrder(order Order) {
	from := board.Regions[order.From]
	from.Order = order
	board.Regions[order.From] = from

	if order.To == "" {
		return
	}

	to := board.Regions[order.To]
	switch order.Type {
	case OrderMove:
		to.IncomingMoves = append(to.IncomingMoves, order)
	case OrderSupport:
		to.IncomingSupports = append(to.IncomingSupports, order)
	}
	board.Regions[order.To] = to
}

// Removes the given move order from the regions on the board.
func (board Board) removeMove(move Order) {
	from := board.Regions[move.From]
	from.Order = Order{}
	board.Regions[move.From] = from

	to := board.Regions[move.To]
	newMoves := make([]Order, 0)
	for _, incMove := range to.IncomingMoves {
		if incMove != move {
			newMoves = append(newMoves, incMove)
		}
	}
	to.IncomingMoves = newMoves
	board.Regions[move.To] = to
}

// Removes the given support order from the regions on the board.
func (board Board) removeSupport(support Order) {
	from := board.Regions[support.From]
	from.Order = Order{}
	board.Regions[support.From] = from

	to := board.Regions[support.To]
	newSupports := make([]Order, 0)
	for _, incSupport := range to.IncomingSupports {
		if incSupport != support {
			newSupports = append(newSupports, incSupport)
		}
	}
	to.IncomingSupports = newSupports
	board.Regions[support.To] = to
}

// Attempts to move the unit of the given move order back to its origin.
// Returns whether the retreat succeeded.
func (board Board) attemptRetreat(move Order) bool {
	from := board.Regions[move.From]

	if from.Unit == move.Unit {
		return true
	}

	if len(from.IncomingMoves) != 0 {
		return false
	}

	from.Unit = move.Unit
	board.Regions[move.From] = from
	return true
}
