package game

import (
	. "immerse-ntnu/hermannia/server/types"
)

func ResolveRound(round *GameRound) {
	activeOrders := resolveOrders(round.Board, round.FirstOrders)
	round.SecondOrders = append(round.SecondOrders, activeOrders...)
	activeOrders = resolveOrders(round.Board, round.SecondOrders)
	resolveFinalOrders(round.Board, activeOrders)
}

func populateAreaOrders(board Board, orders []*Order) {
	for _, order := range orders {
		if to, ok := board[order.To.Name]; ok {
			switch order.Type {
			case Move:
				to.IncomingMoves = append(to.IncomingMoves, order)
			case Support:
				to.IncomingSupports = append(to.IncomingSupports, order)
			}
		}
		if from, ok := board[order.From.Name]; ok {
			from.Outgoing = order
		}
	}
}

func cutSupports(board Board) {
	for _, area := range board {
		if area.Outgoing.Type == Support {
			if len(area.IncomingMoves) > 0 {
				area.Outgoing.Status = Fail
				area.Outgoing.To.IncomingSupports = removeOrder(
					area.Outgoing.To.IncomingSupports,
					area.Outgoing,
				)
			}
		}
	}
}

func resolveOrders(board Board, orders []*Order) []*Order {
	populateAreaOrders(board, orders)
	cutSupports(board)

	conflictFreeResolved := false
	for !conflictFreeResolved {
		conflictFreeResolved = resolveConflictFreeOrders(board)
	}

	resolveTransportOrders(board)

	activeOrders := []*Order{}
	for _, order := range orders {
		if order.Status == Pending {
			activeOrders = append(activeOrders, order)
		}
	}
	return activeOrders
}

func resolveConflictFreeOrders(board Board) bool {
	allResolved := true

	for _, area := range board {
		if area.Unit != nil || len(area.IncomingMoves) != 1 {
			continue
		}

		allResolved = false

		if area.Control == Uncontrolled {
			resolveCombatPvE(area)
		} else {
			succeedMove(area, area.IncomingMoves[0])
		}
	}

	return allResolved
}

func resolveTransportOrders(board Board) {
	for _, area := range board {
		if area.Outgoing.Type != Transport {
			continue
		}

		if len(area.IncomingMoves) == 0 {
			continue
		}

		resolveCombat(area)
	}
}

func resolveFinalOrders(board Board, orders []*Order) {

}
