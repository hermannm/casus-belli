package validation

import (
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/gametypes"
)

// Checks if the given set of orders are valid for the state of the board in the given season.
// Assumes that all orders are from the same player.
func ValidateOrders(
	orders []gametypes.Order, board gametypes.Board, season gametypes.Season,
) error {
	for _, order := range orders {
		err := validateOrder(order, board, season)
		if err != nil {
			return fmt.Errorf("invalid order in region %s: %w", order.From, err)
		}
	}

	err := validateOrderSet(orders, board)
	if err != nil {
		return fmt.Errorf("invalid order set: %w", err)
	}

	return nil
}

func validateOrder(order gametypes.Order, board gametypes.Board, season gametypes.Season) error {
	from := board.Regions[order.From]

	if order.Player != from.Unit.Player {
		return errors.New("must have unit in ordered region")
	}

	switch season {
	case gametypes.SeasonWinter:
		return validateWinterOrder(order, from, board)
	default:
		return validateNonWinterOrder(order, from, board)
	}
}

func validateNonWinterOrder(
	order gametypes.Order, from gametypes.Region, board gametypes.Board,
) error {
	if order.Build != "" {
		return errors.New("build orders can only be placed in winter")
	}

	switch order.Type {
	case gametypes.OrderMove:
		fallthrough
	case gametypes.OrderSupport:
		return validateMoveOrSupport(order, from, board)
	case gametypes.OrderBesiege:
		fallthrough
	case gametypes.OrderTransport:
		return validateBesiegeOrTransport(order, from)
	default:
		return errors.New("invalid order type")
	}
}

func validateMoveOrSupport(
	order gametypes.Order, from gametypes.Region, board gametypes.Board,
) error {
	if order.To == "" {
		return errors.New("moves and supports must have destination")
	}

	to, ok := board.Regions[order.To]
	if !ok {
		return fmt.Errorf("destination region with name %s not found", order.To)
	}

	if from.Unit.Type == gametypes.UnitShip {
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
		return validateMove(order, from, to, board)
	case gametypes.OrderSupport:
		return validateSupport(order, from, to)
	}

	return errors.New("invalid order type")
}

func validateMove(
	order gametypes.Order, from gametypes.Region, to gametypes.Region, board gametypes.Board,
) error {
	if !from.HasNeighbor(order.To) {
		canTransport, _, _ := board.FindTransportPath(from.Name, order.To)
		if !canTransport {
			return errors.New("move is not adjacent to destination, and cannot be transported")
		}
	}

	if from.IsEmpty() || from.Unit.Player != order.Player {
		secondHorseMove := false

		for _, firstOrder := range from.IncomingMoves {
			if from.Unit.Type == gametypes.UnitHorse &&
				order.To == order.From &&
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

func validateSupport(order gametypes.Order, from gametypes.Region, to gametypes.Region) error {
	if !from.HasNeighbor(order.To) {
		return errors.New("support order must be adjacent to destination")
	}

	return nil
}

func validateBesiegeOrTransport(order gametypes.Order, from gametypes.Region) error {
	if order.To != "" {
		return errors.New("besiege or transport orders cannot have destination")
	}

	switch order.Type {
	case gametypes.OrderBesiege:
		return validateBesiege(order, from)
	case gametypes.OrderTransport:
		return validateTransport(order, from)
	default:
		return errors.New("invalid order type")
	}
}

func validateBesiege(order gametypes.Order, from gametypes.Region) error {
	if !from.Castle {
		return errors.New("besieged region must have castle")
	}

	if from.IsControlled() {
		return errors.New("besieged region cannot already be controlled")
	}

	if from.Unit.Type == gametypes.UnitShip {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(order gametypes.Order, from gametypes.Region) error {
	if from.Unit.Type != gametypes.UnitShip {
		return errors.New("only ships can transport")
	}

	if !from.Sea {
		return errors.New("transport orders can only be placed at sea")
	}

	return nil
}

func validateWinterOrder(
	order gametypes.Order, from gametypes.Region, board gametypes.Board,
) error {
	switch order.Type {
	case gametypes.OrderMove:
		return validateWinterMove(order, from, board)
	case gametypes.OrderBuild:
		return validateBuild(order, from, board)
	}

	return errors.New("invalid order type")
}

func validateWinterMove(order gametypes.Order, from gametypes.Region, board gametypes.Board) error {
	if order.To == "" {
		return errors.New("winter move orders must have destination")
	}

	to, ok := board.Regions[order.To]
	if !ok {
		return fmt.Errorf("destination region with name %s not found", order.To)
	}

	if to.ControllingPlayer != order.Player {
		return errors.New("must control destination region in winter move")
	}

	if from.Unit.Type == gametypes.UnitShip && !to.IsCoast(board) {
		return errors.New("ship winter move destination must be coast")
	}

	if order.Build != "" {
		return errors.New("cannot build unit with move order")
	}

	return nil
}

func validateBuild(order gametypes.Order, from gametypes.Region, board gametypes.Board) error {
	if !from.IsEmpty() {
		return errors.New("cannot build in region already occupied")
	}

	switch order.Build {
	case gametypes.UnitShip:
		if !from.IsCoast(board) {
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
			_, notUnique := moveDestinations[order.To]
			if notUnique {
				return fmt.Errorf("orders include two moves to region %s", order.To)
			}

			moveDestinations[order.To] = struct{}{}
		}
	}

	return nil
}

func validateOneOrderPerRegion(orders []gametypes.Order, board gametypes.Board) error {
	orderedRegions := make(map[string]struct{})

	for _, order := range orders {
		_, alreadyOrdered := orderedRegions[order.From]
		if alreadyOrdered {
			return fmt.Errorf("unit in region %s is ordered twice", order.From)
		}

		orderedRegions[order.From] = struct{}{}
	}

	return nil
}
