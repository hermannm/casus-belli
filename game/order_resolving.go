package game

import (
	. "immerse-ntnu/hermannia/server/types"
)

func ResolveOrders(board Board, orders []*Order) {
	firstOrders := []*Order{}
	nextOrders := []*Order{}

	for _, order := range orders {
		if len(order.Dependencies) == 0 {
			firstOrders = append(firstOrders, order)
		} else {
			nextOrders = append(nextOrders, order)
		}
	}

	activeOrders := resolveOrders(board, firstOrders)
	nextOrders = append(nextOrders, activeOrders...)
	activeOrders = resolveOrders(board, nextOrders)
	resolveFinalOrders(board, activeOrders)
}

func populateIncomingOrders(board Board, orders []*Order) {
	for _, area := range board {
		area.Incoming = []*Order{}
	}

	for _, order := range orders {
		if to, ok := board[order.To.Name]; ok {
			to.Incoming = append(to.Incoming, order)
		}
	}
}

func resolveOrders(board Board, orders []*Order) []*Order {
	populateIncomingOrders(board, orders)
	resolveConflictFreeOrders(board, orders)

	activeOrders := []*Order{}
	for _, order := range orders {
		if order.Status == Pending {
			activeOrders = append(activeOrders, order)
		}
	}
	return activeOrders
}

func resolveConflictFreeOrders(board Board, orders []*Order) {
	for _, order := range orders {
		if order.Type != Move {
			continue
		}
		if order.To.Unit != nil {
			continue
		}

		movesToDest := 0
		for _, orderToDest := range order.To.Incoming {
			if orderToDest.Type == Move {
				movesToDest++
			}
		}

		if movesToDest == 1 {
			if order.To.Control == Uncontrolled {
				resolveCombat(order.To)
			} else {
				succeedMove(order)
			}
		}
	}
}

func resolveFinalOrders(board Board, orders []*Order) {

}

func succeedMove(order *Order) {
	order.To.Control = order.Player.Color
	order.To.Unit = order.From.Unit
	order.From.Unit = nil
	order.Status = Success
}
