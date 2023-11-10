package gametypes

import "hermannm.dev/set"

type Board struct {
	// Maps region names to regions.
	Regions            map[string]Region
	Name               string
	WinningCastleCount int
}

func (board Board) RemoveUnit(unit Unit, regionName string) {
	region := board.Regions[regionName]

	if unit == region.Unit {
		region.Unit = Unit{}
		board.Regions[regionName] = region
	}
}

// Populates regions on the board with the given orders.
// Does not add support orders that have moves against them, as that cancels them.
func (board Board) AddOrders(orders []Order) {
	var supportOrders []Order

	for _, order := range orders {
		if order.Type == OrderSupport {
			supportOrders = append(supportOrders, order)
			continue
		}

		board.AddOrder(order)
	}

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

func (board Board) RemoveOrders() {
	for regionName, region := range board.Regions {
		region.Order = Order{}
		region.IncomingMoves = nil
		region.IncomingSupports = nil

		board.Regions[regionName] = region
	}
}

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

func (board Board) CheckWinner() (winner PlayerFaction) {
	castleCount := make(map[PlayerFaction]int)

	for _, region := range board.Regions {
		if region.HasCastle && region.IsControlled() {
			castleCount[region.ControllingFaction]++
		}
	}

	tie := false
	highestCount := 0
	var highestCountFaction PlayerFaction
	for faction, count := range castleCount {
		if count > highestCount {
			highestCount = count
			highestCountFaction = faction
			tie = false
		} else if count == highestCount {
			tie = true
		}
	}

	if tie || highestCount < board.WinningCastleCount {
		return ""
	}

	return highestCountFaction
}

// Returns a list of the factions that players can use on this board.
func (board Board) AvailablePlayerFactions() []PlayerFaction {
	var factions set.ArraySet[PlayerFaction]

	for _, region := range board.Regions {
		if region.HomeFaction != "" {
			factions.Add(region.HomeFaction)
		}
	}

	return factions.ToSlice()
}
