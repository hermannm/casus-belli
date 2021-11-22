package game

import (
	. "immerse-ntnu/hermannia/server/types"
)

func ResolveOrders(board *Board, orders []*Order) {
	var firstOrders []*Order
	var nextOrders []*Order

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

func makeIncomingOrdersMap(board Board, orders []*Order) map[string][]*Order {
	incomingOrders := make(map[string][]*Order)

	for areaName, _ := range board {
		incomingOrders[areaName] = make([]*Order, 0)
	}

	for _, order := range orders {
		if to, ok := incomingOrders[order.To.Name]; ok {
			to = append(to, order)
		}
	}

	return incomingOrders
}

func resolveOrders(board *Board, orders []*Order) []*Order {
	incomingOrders := makeIncomingOrdersMap(*board, orders)

	for areaName, area := range *board {
		incoming := incomingOrders[areaName]
	}

	activeOrders := make([]*Order, 0)

	for _, order := range orders {
		if order.Result.Status == Pending {
			activeOrders = append(activeOrders, order)
		}
	}

	return activeOrders
}

func resolveConflictFreeOrders(board *Board, orders []*Order) {

}

func resolveFinalOrders(board *Board, orders []*Order) {

}
