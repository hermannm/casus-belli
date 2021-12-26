package game

// Fails a support order and cleans up references to it on the board.
func (support *Order) failSupport() {
	support.Status = Fail

	support.To.IncomingSupports = removeOrder(
		support.To.IncomingSupports,
		support,
	)

	support.From.Order = nil
}

// Succeeds a move order and adjusts board areas accordingly.
func (move *Order) moveAndSucceed() {
	moveUnit(move.From, move.To)
	move.succeedMove()
}

func (move *Order) succeedMove() {
	move.Status = Success
	move.From.Order = nil
	move.To.IncomingMoves = removeOrder(move.To.IncomingMoves, move)

	// Seas cannot be controlled.
	if !move.To.Sea {
		move.To.Control = move.Player
	}
}

// Fails a move order and cleans up references to it on the board.
func (move *Order) failMove() {
	move.Status = Fail
	move.From.Order = nil
	move.To.IncomingMoves = removeOrder(move.To.IncomingMoves, move)
}

// Removes a move order's unit from the board.
func (move *Order) killAttacker() {
	move.From.Unit = Unit{}
}

// Succeeds move from the winner after battle in an area,
// and fails all others.
func (area *BoardArea) resolveWinner(winner Player) {
	if !area.IsEmpty() && area.Unit.Player != winner {
		area.removeUnit()
	}

	for _, move := range area.IncomingMoves {
		if move.Player == winner {
			move.moveAndSucceed()
		} else {
			move.failMove()
			move.killAttacker()
		}
	}
}

// Fails all units involved in battle except winner.
// Used for resolving battles that have follow-up battles,
// and so should not succeed the winning move yet.
func (area *BoardArea) resolveIntermediaryWinner(winner Player) {
	if !area.IsEmpty() && area.Unit.Player != winner {
		area.removeUnit()
	}

	for _, move := range area.IncomingMoves {
		if move.Player != winner {
			move.failMove()
			move.killAttacker()
		}
	}
}

// Rolls dice to see if order makes it across danger zone.
// Returns true if order succeeded.
// If order is a move and fails, it is killed.
// Adds result to battle list of origin area.
func (order *Order) crossDangerZone() bool {
	diceMod := diceModifier()

	// Records crossing attempt as a battle, so clients can see dice roll.
	battle := Battle{
		{
			Total:  diceMod.Value,
			Parts:  []Modifier{diceMod},
			Player: order.Player,
		},
	}
	order.From.Battles = append(order.From.Battles, battle)

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

// Takes a list of order references and returns it without the removed order.
func removeOrder(oldOrders []*Order, remove *Order) []*Order {
	newOrders := make([]*Order, 0)

	for _, order := range oldOrders {
		if order != remove {
			newOrders = append(newOrders, order)
		}
	}

	return newOrders
}
