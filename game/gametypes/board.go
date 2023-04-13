package gametypes

// A pre-configured board used for the game.
type Board struct {
	// Regions on the board that player units can move to.
	Regions map[string]Region `json:"region"`

	// Name of this type of board.
	Name string `json:"name"`

	// The number of castles to capture to win a game round on this board.
	WinningCastleCount int `json:"winningCastleCount"`
}

func (board Board) RegionNames() []string {
	regionNames := make([]string, 0, len(board.Regions))

	for regionName := range board.Regions {
		regionNames = append(regionNames, regionName)
	}

	return regionNames
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
	var supportOrders []Order

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
		if !board.Regions[supportOrder.Origin].IsAttacked() {
			board.AddOrder(supportOrder)
		}
	}
}

func (board Board) AddOrder(order Order) {
	origin := board.Regions[order.Origin]
	origin.Order = order
	board.Regions[order.Origin] = origin

	if order.Destination == "" {
		return
	}

	destination := board.Regions[order.Destination]
	switch order.Type {
	case OrderMove:
		destination.IncomingMoves = append(destination.IncomingMoves, order)
	case OrderSupport:
		destination.IncomingSupports = append(destination.IncomingSupports, order)
	}
	board.Regions[order.Destination] = destination
}

// Removes the given order from the regions on the board.
func (board Board) RemoveOrder(order Order) {
	origin := board.Regions[order.Origin]
	origin.Order = Order{}
	board.Regions[order.Origin] = origin

	switch order.Type {
	case OrderMove:
		destination := board.Regions[order.Destination]

		var newMoves []Order
		for _, incomingMove := range destination.IncomingMoves {
			if incomingMove != order {
				newMoves = append(newMoves, incomingMove)
			}
		}
		destination.IncomingMoves = newMoves

		board.Regions[order.Destination] = destination
	case OrderSupport:
		destination := board.Regions[order.Destination]

		var newSupports []Order
		for _, incSupport := range destination.IncomingSupports {
			if incSupport != order {
				newSupports = append(newSupports, incSupport)
			}
		}
		destination.IncomingSupports = newSupports

		board.Regions[order.Destination] = destination
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
