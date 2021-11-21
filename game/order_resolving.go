package game

import (
	"immerse-ntnu/hermannia/server/types"
)

func (board *board) ResolveOrders(orders []types.Order) []types.OrderResult {
	var results []types.OrderResult

	for _, order := range orders {
		results = append(results, types.OrderResult{
			Status: types.Success,
			Order:  order,
		})
	}

	return results
}
