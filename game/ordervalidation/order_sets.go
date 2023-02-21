package ordervalidation

import (
	"fmt"

	"hermannm.dev/bfh-server/game/gametypes"
)

func validateOrderSet(orders []gametypes.Order, board gametypes.Board) error {
	err := validateUniqueMoveDestinations(orders, board)
	if err != nil {
		return err
	}

	err = validateOneOrderPerRegion(orders, board)
	if err != nil {
		return err
	}

	return nil
}

func validateUniqueMoveDestinations(orders []gametypes.Order, board gametypes.Board) error {
	moveDestinations := make(map[string]struct{})

	for _, order := range orders {
		if order.Type == gametypes.OrderMove {
			_, notUnique := moveDestinations[order.Destination]
			if notUnique {
				return fmt.Errorf("orders include two moves to region %s", order.Destination)
			}

			moveDestinations[order.Destination] = struct{}{}
		}
	}

	return nil
}

func validateOneOrderPerRegion(orders []gametypes.Order, board gametypes.Board) error {
	orderedRegions := make(map[string]struct{})

	for _, order := range orders {
		_, alreadyOrdered := orderedRegions[order.Origin]
		if alreadyOrdered {
			return fmt.Errorf("unit in region %s is ordered twice", order.Origin)
		}

		orderedRegions[order.Origin] = struct{}{}
	}

	return nil
}
