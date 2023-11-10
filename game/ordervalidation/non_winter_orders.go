package ordervalidation

import (
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/wrap"
)

func validateNonWinterOrders(orders []gametypes.Order, board gametypes.Board) error {
	for _, order := range orders {
		origin, ok := board.Regions[order.Origin]
		if !ok {
			return wrap.Error(
				fmt.Errorf("origin region with name '%s' not found", order.Origin), "invalid order",
			)
		}

		if err := validateNonWinterOrder(order, origin, board); err != nil {
			return wrap.Errorf(err, "invalid order in region '%s'", order.Origin)
		}
	}

	if err := validateOrderSet(orders, board); err != nil {
		return wrap.Error(err, "invalid order set")
	}

	if err := validateReachableMoveDestinations(orders, board); err != nil {
		return err
	}

	return nil
}

func validateNonWinterOrder(
	order gametypes.Order, origin gametypes.Region, board gametypes.Board,
) error {
	if order.Build != "" {
		return errors.New("build orders can only be placed in winter")
	}

	if order.Faction != origin.Unit.Faction {
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
		return fmt.Errorf("invalid order type '%s'", order.Type)
	}
}

func validateMoveOrSupport(
	order gametypes.Order, origin gametypes.Region, board gametypes.Board,
) error {
	if order.Destination == "" {
		return errors.New("moves and supports must have destination")
	}

	destination, ok := board.Regions[order.Destination]
	if !ok {
		return fmt.Errorf("destination region with name '%s' not found", order.Destination)
	}

	if origin.Unit.Type == gametypes.UnitShip {
		if !(destination.IsSea || destination.IsCoast(board)) {
			return errors.New("ship order destination must be sea or coast")
		}
	} else {
		if destination.IsSea {
			return errors.New("only ships can order to seas")
		}
	}

	switch order.Type {
	case gametypes.OrderMove:
		return validateMove(order, origin, board)
	case gametypes.OrderSupport:
		return validateSupport(order, origin, destination)
	}

	return errors.New("invalid order type")
}

func validateMove(order gametypes.Order, origin gametypes.Region, board gametypes.Board) error {
	if order.SecondDestination != "" {
		if origin.Unit.Type != gametypes.UnitHorse {
			return errors.New(
				"second destinations for move orders can only be applied to horse units",
			)
		}

		if _, ok := board.Regions[order.SecondDestination]; !ok {
			return fmt.Errorf(
				"second destination region with name '%s' not found", order.SecondDestination,
			)
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
	if !origin.HasCastle {
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

	if !origin.IsSea {
		return errors.New("transport orders can only be placed at sea")
	}

	return nil
}

func validateReachableMoveDestinations(orders []gametypes.Order, board gametypes.Board) error {
	boardCopy := gametypes.Board{Regions: make(map[string]gametypes.Region, len(board.Regions))}
	for regionName, region := range board.Regions {
		boardCopy.Regions[regionName] = region
	}

	boardCopy.AddOrders(orders)

	for _, order := range orders {
		if order.Type != gametypes.OrderMove {
			continue
		}

		if err := validateReachableMoveDestination(order, boardCopy); err != nil {
			return wrap.Errorf(
				err, "invalid move from '%s' to '%s'", order.Origin, order.Destination,
			)
		}

		secondHorseMove, hasSecondHorseMove := order.TryGetSecondHorseMove()
		if hasSecondHorseMove {
			if err := validateReachableMoveDestination(secondHorseMove, boardCopy); err != nil {
				return wrap.Errorf(
					err,
					"invalid second destination for horse move from '%s' to '%s'",
					order.Origin,
					order.SecondDestination,
				)
			}
		}
	}

	return nil
}

func validateReachableMoveDestination(move gametypes.Order, board gametypes.Board) error {
	origin := board.Regions[move.Origin]

	if !origin.HasNeighbor(move.Destination) {
		canTransport, _, _ := board.FindTransportPath(move.Origin, move.Destination)

		if !canTransport {
			return errors.New("regions not adjacent, and no transport path available")
		}
	}

	return nil
}
