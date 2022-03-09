package board

// Checks whether the order is initialized.
func (order Order) IsNone() bool {
	return order.Type == ""
}

// Moves the unit of the given move order to its destination,
// killing any unit that may have already been there,
// and sets control of the area to the order's player.
//
// Then removes references to this move on the board,
// and removes any potential order from the destination area.
func (board Board) succeedMove(move Order) {
	to := board[move.To]

	to = to.setUnit(move.Unit)
	to = to.setOrder(Order{})
	if !to.Sea {
		to = to.setControl(move.Player)
	}
	board[move.To] = to

	board.removeOriginUnit(move)

	board.removeMove(move)
}

// Removes the move order's unit from its origin area, if it still exists.
func (board Board) removeOriginUnit(move Order) {
	from := board[move.From]

	if move.Unit == from.Unit {
		board[move.From] = from.setUnit(Unit{})
	}
}

// Rolls dice to see if order makes it across danger zone.
// Returns whether the order succeeded, and the resulting battle for use by the client.
func (order Order) crossDangerZone(dangerZone string) (survived bool, result Battle) {
	diceMod := Modifier{
		Type:  ModifierDice,
		Value: rollDice(),
	}

	// Records crossing attempt as a battle, so clients can see dice roll.
	battle := Battle{
		Results: []Result{
			{
				Total: diceMod.Value,
				Parts: []Modifier{diceMod},
				Move:  order,
			},
		},
		DangerZone: dangerZone,
	}

	// All danger zones currently require a dice roll greater than 2.
	// May need to be changed in the future if a more dynamic implementation is preferred.
	return diceMod.Value > 2, battle
}

func (board Board) addOrder(order Order) {
	board[order.From] = board[order.From].setOrder(order)

	if order.To == "" {
		return
	}

	to := board[order.To]
	switch order.Type {
	case OrderMove:
		to.IncomingMoves = append(to.IncomingMoves, order)
	case OrderSupport:
		to.IncomingSupports = append(to.IncomingSupports, order)
	}
	board[order.To] = to
}

// Removes the given move order from the areas on the board.
func (board Board) removeMove(move Order) {
	board[move.From] = board[move.From].setOrder(Order{})
	board[move.To] = board[move.To].removeIncomingMove(move)
}

// Returns the given area with the given order removed from its list of incoming moves.
// Assumes the given order is a move order.
func (area Area) removeIncomingMove(move Order) Area {
	newMoves := make([]Order, 0)
	for _, incMove := range area.IncomingMoves {
		if incMove != move {
			newMoves = append(newMoves, incMove)
		}
	}
	area.IncomingMoves = newMoves
	return area
}

// Removes the given support order from the areas on the board.
func (board Board) removeSupport(support Order) {
	board[support.From] = board[support.From].setOrder(Order{})
	board[support.To] = board[support.To].removeIncomingSupport(support)
}

// Returns the given area with the given order removed from its list of incoming supports.
// Assumes the given order is a support order.
func (area Area) removeIncomingSupport(support Order) Area {
	newSupports := make([]Order, 0)
	for _, incSupport := range area.IncomingSupports {
		if incSupport != support {
			newSupports = append(newSupports, incSupport)
		}
	}
	area.IncomingSupports = newSupports
	return area
}

// Attempts to move the unit of the given move order back to its origin.
// Returns whether the retreat succeeded.
func (board Board) attemptRetreat(move Order) bool {
	from := board[move.From]

	if from.Unit == move.Unit {
		return true
	}

	if len(from.IncomingMoves) != 0 {
		return false
	}

	board[move.From] = from.setUnit(move.Unit)
	return true
}
