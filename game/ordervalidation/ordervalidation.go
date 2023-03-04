package ordervalidation

import (
	"hermannm.dev/bfh-server/game/gametypes"
)

// Checks if the given set of orders are valid for the state of the board in the given season.
// Assumes that all orders are from the same player.
func ValidateOrders(
	orders []gametypes.Order, board gametypes.Board, season gametypes.Season,
) error {
	var err error
	if season == gametypes.SeasonWinter {
		err = validateWinterOrders(orders, board)
	} else {
		err = validateNonWinterOrders(orders, board)
	}
	return err
}
