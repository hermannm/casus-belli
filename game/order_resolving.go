package game

func removeOrder(oldOrders []*Order, remove *Order) []*Order {
	newOrders := make([]*Order, 0)

	for _, order := range oldOrders {
		if order != remove {
			newOrders = append(newOrders, order)
		}
	}

	return newOrders
}

func (support *Order) failSupport() {
	support.Status = Fail

	support.To.IncomingSupports = removeOrder(
		support.To.IncomingSupports,
		support,
	)

	support.From.Outgoing = nil
}

func (order *Order) succeedMove() {
	order.To.Unit = order.From.Unit
	order.Status = Success
	order.From.Unit = nil
	order.From.Outgoing = nil
	order.To.IncomingMoves = removeOrder(order.To.IncomingMoves, order)

	// Seas cannot be controlled.
	if !order.To.Sea {
		order.To.Control = order.Player
	}
}

func (order *Order) failMove() {
	order.Status = Fail
	order.From.Outgoing = nil
	order.To.IncomingMoves = removeOrder(order.To.IncomingMoves, order)
}

func (order *Order) die() {
	order.From.Unit = nil
}

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
			move.die()
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
			move.die()
		}
	}
}

// Rolls dice to see if order makes it across danger zone.
// If order is a move and fails, it is killed.
// Adds result to combat list of origin area.
// Returns true if order succeeded.
func (order *Order) crossDangerZone() bool {
	diceMod := DiceModifier()

	combat := Combat{
		{
			Total:  diceMod.Value,
			Parts:  []Modifier{diceMod},
			Player: order.Player,
		},
	}

	// Records crossing attempt as a combat, so clients can see dice roll.
	order.From.Combats = append(order.From.Combats, combat)

	// All danger zones currently require a dice roll greater than 2.
	// May need to be changed in the future if a more dynamic implementation is preferred.
	if diceMod.Value <= 2 {
		switch order.Type {
		case Move:
			order.failMove()
			order.die()
		case Support:
			order.failSupport()
		}

		return false
	}
	return true
}
