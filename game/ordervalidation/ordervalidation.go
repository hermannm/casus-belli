package ordervalidation

import (
	"fmt"

	"hermannm.dev/bfh-server/game/gametypes"
)

// Checks if the given set of orders are valid for the state of the board in the given season.
// Assumes that all orders are from the same player.
func ValidateOrders(
	orders []gametypes.Order, board gametypes.Board, season gametypes.Season,
) error {
	for _, order := range orders {
		origin := board.Regions[order.Origin]

		var err error
		if season == gametypes.SeasonWinter {
			err = validateWinterOrder(order, origin, board)
		} else {
			err = validateNonWinterOrder(order, origin, board)
		}

		if err != nil {
			return fmt.Errorf("invalid order in region %s: %w", order.Origin, err)
		}
	}

	err := validateOrderSet(orders, board)
	if err != nil {
		return fmt.Errorf("invalid order set: %w", err)
	}

	return nil
}
