package game

// Fails a support order and cleans up references to it on the board.
func (support *Order) failSupport() {
	support.Status = Fail

	support.To.IncomingSupports = removeOrder(
		support.To.IncomingSupports,
		support,
	)

	support.From.Outgoing = nil
}

// Succeeds a move order and adjusts board areas accordingly.
func (move *Order) succeedMove() {
	move.To.Unit = move.From.Unit
	move.Status = Success
	move.From.Unit = nil
	move.From.Outgoing = nil
	move.To.IncomingMoves = removeOrder(move.To.IncomingMoves, move)

	// Seas cannot be controlled.
	if !move.To.Sea {
		move.To.Control = move.Player
	}
}

// Fails a move order and cleans up references to it on the board.
func (move *Order) failMove() {
	move.Status = Fail
	move.From.Outgoing = nil
	move.To.IncomingMoves = removeOrder(move.To.IncomingMoves, move)
}

// Removes a move order's unit from the board.
func (move *Order) killAttacker() {
	move.From.Unit = nil
}

// Removes a defending unit from the board, and fails its order if any.
func (area *BoardArea) killDefender() {
	area.Unit = nil
	if area.Outgoing != nil {
		area.Outgoing.Status = Fail
		area.Outgoing = nil
	}
}

// Succeeds move from the winner after combat in an area,
// and fails all others.
func (area *BoardArea) resolveWinner(winner PlayerColor) {
	if area.Unit != nil && area.Unit.Color != winner {
		area.killDefender()
	}

	for _, move := range area.IncomingMoves {
		if move.Player == winner {
			move.succeedMove()
		} else {
			move.failMove()
			move.killAttacker()
		}
	}
}

// Fails all units involved in combat except winner.
// Used for resolving combats that have follow-up combats,
// and so should not succeed the winning move yet.
func (area *BoardArea) resolveIntermediaryWinner(winner PlayerColor) {
	if area.Unit != nil && area.Unit.Color != winner {
		area.killDefender()
	}

	for _, move := range area.IncomingMoves {
		if move.Player != winner {
			move.failMove()
			move.killAttacker()
		}
	}
}

// Rolls dice to see if order makes it across danger zone.
// If order is a move and fails, it is killed.
// Adds result to combat list of origin area.
// Returns true if order succeeded.
func (order *Order) crossDangerZone() bool {
	diceMod := diceModifier()

	// Records crossing attempt as a combat, so clients can see dice roll.
	combat := Combat{
		{
			Total:  diceMod.Value,
			Parts:  []Modifier{diceMod},
			Player: order.Player,
		},
	}
	order.From.Combats = append(order.From.Combats, combat)

	// All danger zones currently require a dice roll greater than 2.
	// May need to be changed in the future if a more dynamic implementation is preferred.
	if diceMod.Value <= 2 {
		switch order.Type {
		case Move:
			order.failMove()
			order.killAttacker()
		case Support:
			order.failSupport()
		}

		return false
	}
	return true
}

// Takes a list of orders and returns it without the removed order.
func removeOrder(oldOrders []*Order, remove *Order) []*Order {
	newOrders := make([]*Order, 0)

	for _, order := range oldOrders {
		if order != remove {
			newOrders = append(newOrders, order)
		}
	}

	return newOrders
}
