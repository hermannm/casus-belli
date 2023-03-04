package ordervalidation

import (
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/gametypes"
)

func validateWinterOrders(orders []gametypes.Order, board gametypes.Board) error {
	for _, order := range orders {
		origin, ok := board.Regions[order.Origin]
		if !ok {
			return fmt.Errorf("invalid order: origin region with name '%s' not found", order.Origin)
		}

		if err := validateWinterOrder(order, origin, board); err != nil {
			return fmt.Errorf("invalid winter order in region %s: %w", order.Origin, err)
		}
	}

	if err := validateOrderSet(orders, board); err != nil {
		return fmt.Errorf("invalid winter order set: %w", err)
	}

	return nil
}

func validateWinterOrder(
	order gametypes.Order, origin gametypes.Region, board gametypes.Board,
) error {
	switch order.Type {
	case gametypes.OrderMove:
		return validateWinterMove(order, origin, board)
	case gametypes.OrderBuild:
		return validateBuild(order, origin, board)
	default:
		return fmt.Errorf("invalid order type in winter: %s", order.Type)
	}
}

func validateWinterMove(
	order gametypes.Order, origin gametypes.Region, board gametypes.Board,
) error {
	if order.Destination == "" {
		return errors.New("winter move orders must have destination")
	}

	to, ok := board.Regions[order.Destination]
	if !ok {
		return fmt.Errorf("destination region with name %s not found", order.Destination)
	}

	if to.ControllingPlayer != order.Player {
		return errors.New("must control destination region in winter move")
	}

	if origin.Unit.Type == gametypes.UnitShip && !to.IsCoast(board) {
		return errors.New("ship winter move destination must be coast")
	}

	if order.Build != "" {
		return errors.New("cannot build unit with move order")
	}

	return nil
}

func validateBuild(order gametypes.Order, origin gametypes.Region, board gametypes.Board) error {
	if !origin.IsEmpty() {
		return errors.New("cannot build in region already occupied")
	}

	switch order.Build {
	case gametypes.UnitShip:
		if !origin.IsCoast(board) {
			return errors.New("ships can only be built on coast")
		}
	case gametypes.UnitFootman:
	case gametypes.UnitHorse:
	case gametypes.UnitCatapult:
	default:
		return errors.New("invalid unit type")
	}

	return nil
}
