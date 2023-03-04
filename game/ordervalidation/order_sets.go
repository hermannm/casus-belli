package ordervalidation

import (
	"fmt"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/set"
)

func validateOrderSet(orders []gametypes.Order, board gametypes.Board) error {
	if err := validateUniqueMoveDestinations(orders, board); err != nil {
		return err
	}

	if err := validateOneOrderPerRegion(orders, board); err != nil {
		return err
	}

	return nil
}

func validateUniqueMoveDestinations(orders []gametypes.Order, board gametypes.Board) error {
	moveDestinations := set.New[string]()

	for _, order := range orders {
		if order.Type == gametypes.OrderMove {
			if moveDestinations.Contains(order.Destination) {
				return fmt.Errorf("orders include two moves to region %s", order.Destination)
			}

			if order.SecondDestination != "" && moveDestinations.Contains(order.SecondDestination) {
				return fmt.Errorf("orders include two moves to region %s", order.SecondDestination)
			}

			moveDestinations.Add(order.Destination)
		}
	}

	return nil
}

func validateOneOrderPerRegion(orders []gametypes.Order, board gametypes.Board) error {
	orderedRegions := set.New[string]()

	for _, order := range orders {
		if orderedRegions.Contains(order.Origin) {
			return fmt.Errorf("unit in region %s is ordered twice", order.Origin)
		}

		orderedRegions.Add(order.Origin)
	}

	return nil
}
