package ordervalidation

import (
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/gametypes"
)

func validateNonWinterOrder(
	order gametypes.Order, origin gametypes.Region, board gametypes.Board,
) error {
	if order.Build != "" {
		return errors.New("build orders can only be placed in winter")
	}

	if order.Player != origin.Unit.Player {
		return errors.New("must have unit in ordered region")
	}

	switch order.Type {
	case gametypes.OrderMove:
		fallthrough
	case gametypes.OrderSupport:
		return validateMoveOrSupport(order, origin, board)
	case gametypes.OrderBesiege:
		fallthrough
	case gametypes.OrderTransport:
		return validateBesiegeOrTransport(order, origin)
	default:
		return errors.New("invalid order type")
	}
}

func validateMoveOrSupport(
	order gametypes.Order, origin gametypes.Region, board gametypes.Board,
) error {
	if order.Destination == "" {
		return errors.New("moves and supports must have destination")
	}

	to, ok := board.Regions[order.Destination]
	if !ok {
		return fmt.Errorf("destination region with name %s not found", order.Destination)
	}

	if origin.Unit.Type == gametypes.UnitShip {
		if !(to.Sea || to.IsCoast(board)) {
			return errors.New("ship order destination must be sea or coast")
		}
	} else {
		if to.Sea {
			return errors.New("only ships can order to seas")
		}
	}

	switch order.Type {
	case gametypes.OrderMove:
		return validateMove(order, origin, to, board)
	case gametypes.OrderSupport:
		return validateSupport(order, origin, to)
	}

	return errors.New("invalid order type")
}

func validateMove(
	order gametypes.Order,
	origin gametypes.Region,
	destination gametypes.Region,
	board gametypes.Board,
) error {
	if !origin.HasNeighbor(order.Destination) {
		canTransport, _, _ := board.FindTransportPath(origin.Name, order.Destination)
		if !canTransport {
			return errors.New("move is not adjacent to destination, and cannot be transported")
		}
	}

	if origin.IsEmpty() || origin.Unit.Player != order.Player {
		secondHorseMove := false

		for _, firstOrder := range origin.IncomingMoves {
			if origin.Unit.Type == gametypes.UnitHorse &&
				order.Destination == order.Origin &&
				firstOrder.Player == order.Player {

				secondHorseMove = true
			}
		}

		if !secondHorseMove {
			return errors.New("must have unit in origin region")
		}
	}

	return nil
}

func validateSupport(
	order gametypes.Order, origin gametypes.Region, destination gametypes.Region,
) error {
	if !origin.HasNeighbor(order.Destination) {
		return errors.New("support order must be adjacent to destination")
	}

	return nil
}

func validateBesiegeOrTransport(order gametypes.Order, origin gametypes.Region) error {
	if order.Destination != "" {
		return errors.New("besiege or transport orders cannot have destination")
	}

	switch order.Type {
	case gametypes.OrderBesiege:
		return validateBesiege(order, origin)
	case gametypes.OrderTransport:
		return validateTransport(order, origin)
	default:
		return errors.New("invalid order type")
	}
}

func validateBesiege(order gametypes.Order, origin gametypes.Region) error {
	if !origin.Castle {
		return errors.New("besieged region must have castle")
	}

	if origin.IsControlled() {
		return errors.New("besieged region cannot already be controlled")
	}

	if origin.Unit.Type == gametypes.UnitShip {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(order gametypes.Order, origin gametypes.Region) error {
	if origin.Unit.Type != gametypes.UnitShip {
		return errors.New("only ships can transport")
	}

	if !origin.Sea {
		return errors.New("transport orders can only be placed at sea")
	}

	return nil
}
