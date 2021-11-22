package game

import (
	. "immerse-ntnu/hermannia/server/types"
)

func (board *board) ResolveOrders(orders []*Order) []OrderResult {
	var results []OrderResult

	for _, order := range orders {
		results = append(results, OrderResult{
			Status: Success,
			Order:  order,
		})
	}

	return results
}
