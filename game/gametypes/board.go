package gametypes

// A pre-configured board used for the game.
type Board struct {
	// Regions on the board that player units can move to.
	Regions map[string]Region `json:"region"`

	// Name of this type of board.
	Name string

	// The number of castles to capture to win a game round on this board.
	WinningCastleCount int `json:"winningCastleCount"`
}

// Removes the given unit from the region with the given name, if the unit still exists there.
func (board Board) RemoveUnit(unit Unit, regionName string) {
	region := board.Regions[regionName]

	if unit == region.Unit {
		region.Unit = Unit{}
		board.Regions[regionName] = region
	}
}

// Takes a list of orders, and populates the appropriate regions on the board with those orders.
// Does not add support orders that have moves against them, as that cancels them.
func (board Board) AddOrders(orders []Order) {
	supportOrders := make([]Order, 0, len(orders))

	// First adds all orders except supports, so that supports can check IncomingMoves.
	for _, order := range orders {
		if order.Type == OrderSupport {
			supportOrders = append(supportOrders, order)
			continue
		}

		board.AddOrder(order)
	}

	// Then adds all supports, except in those regions that are attacked.
	for _, supportOrder := range supportOrders {
		if len(board.Regions[supportOrder.From].IncomingMoves) == 0 {
			board.AddOrder(supportOrder)
		}
	}
}

func (board Board) AddOrder(order Order) {
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

// Cleans up remaining order references on the board after the round.
func (board Board) RemoveOrders() {
	for regionName, region := range board.Regions {
		region.Order = Order{}

		if len(region.IncomingMoves) > 0 {
			region.IncomingMoves = make([]Order, 0)
		}
		if len(region.IncomingSupports) > 0 {
			region.IncomingSupports = make([]Order, 0)
		}

		board.Regions[regionName] = region
	}
}

// Removes the given order from the regions on the board.
func (board Board) RemoveOrder(order Order) {
	from := board.Regions[order.From]
	from.Order = Order{}
	board.Regions[order.From] = from

	switch order.Type {
	case OrderMove:
		to := board.Regions[order.To]

		newMoves := make([]Order, 0)
		for _, incMove := range to.IncomingMoves {
			if incMove != order {
				newMoves = append(newMoves, incMove)
			}
		}
		to.IncomingMoves = newMoves

		board.Regions[order.To] = to
	case OrderSupport:
		to := board.Regions[order.To]

		newSupports := make([]Order, 0)
		for _, incSupport := range to.IncomingSupports {
			if incSupport != order {
				newSupports = append(newSupports, incSupport)
			}
		}
		to.IncomingSupports = newSupports

		board.Regions[order.To] = to
	}
}

// Goes through the board to check if any player has met the board's winning castle count.
// If there is a winner, and there is no tie, returns the tag of that player.
// Otherwise, returns hasWinner=false.
func (board Board) CheckWinner() (winner string, hasWinner bool) {
	castleCount := make(map[string]int)

	for _, region := range board.Regions {
		if region.Castle && region.IsControlled() {
			castleCount[region.ControllingPlayer]++
		}
	}

	tie := false
	highestCount := 0
	var highestCountPlayer string
	for player, count := range castleCount {
		if count > highestCount {
			highestCount = count
			highestCountPlayer = player
			tie = false
		} else if count == highestCount {
			tie = true
		}
	}

	hasWinner = !tie && highestCount > board.WinningCastleCount
	return highestCountPlayer, hasWinner
}
