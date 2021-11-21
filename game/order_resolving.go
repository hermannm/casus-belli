package game

import (
	t "immerse-ntnu/hermannia/server/types"
)

func (board *board) ResolveOrders(orders []*t.Order) []t.OrderResult {
	var results []t.OrderResult

	for _, order := range orders {
		results = append(results, t.OrderResult{
			Status: t.Success,
			Order:  order,
		})
	}

	return results
}
