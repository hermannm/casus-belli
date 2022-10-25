package validation

import (
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/board"
)

// Checks if the given set of orders are valid for the state of the board in the given season.
// Assumes that all orders are from the same player.
func ValidateOrders(orders []board.Order, brd board.Board, season board.Season) error {
	for _, order := range orders {
		if err := validateOrder(order, brd, season); err != nil {
			return fmt.Errorf("invalid order in area %s: %w", order.From, err)
		}
	}

	if err := validateOrderSet(orders, brd); err != nil {
		return fmt.Errorf("invalid order set: %w", err)
	}

	return nil
}

func validateOrder(order board.Order, brd board.Board, season board.Season) error {
	from := brd.Areas[order.From]

	if order.Player != from.Unit.Player {
		return errors.New("must have unit in ordered area")
	}

	switch season {
	case board.SeasonWinter:
		return validateWinterOrder(order, from, brd)
	default:
		return validateNonWinterOrder(order, from, brd)
	}
}

func validateNonWinterOrder(order board.Order, from board.Area, brd board.Board) error {
	if order.Build != "" {
		return errors.New("build orders can only be placed in winter")
	}

	switch order.Type {
	case board.OrderMove:
		fallthrough
	case board.OrderSupport:
		return validateMoveOrSupport(order, from, brd)
	case board.OrderBesiege:
		fallthrough
	case board.OrderTransport:
		return validateBesiegeOrTransport(order, from)
	default:
		return errors.New("invalid order type")
	}
}

func validateMoveOrSupport(order board.Order, from board.Area, brd board.Board) error {
	if order.To == "" {
		return errors.New("moves and supports must have destination")
	}

	to, ok := brd.Areas[order.To]
	if !ok {
		return fmt.Errorf("destination area with name %s not found", order.To)
	}

	if from.Unit.Type == board.UnitShip {
		if !(to.Sea || to.IsCoast(brd)) {
			return errors.New("ship order destination must be sea or coast")
		}
	} else {
		if to.Sea {
			return errors.New("only ships can order to seas")
		}
	}

	switch order.Type {
	case board.OrderMove:
		return validateMove(order, from, to, brd)
	case board.OrderSupport:
		return validateSupport(order, from, to)
	}

	return errors.New("invalid order type")
}

func validateMove(order board.Order, from board.Area, to board.Area, brd board.Board) error {
	if !from.HasNeighbor(order.To) {
		if canTransport, _, _ := from.CanTransportTo(order.To, brd); !canTransport {
			return errors.New("move is not adjacent to destination, and cannot be transported")
		}
	}

	if from.IsEmpty() || from.Unit.Player != order.Player {
		secondHorseMove := false

		for _, firstOrder := range from.IncomingMoves {
			if from.Unit.Type == board.UnitHorse &&
				order.To == order.From &&
				firstOrder.Player == order.Player {

				secondHorseMove = true
			}
		}

		if !secondHorseMove {
			return errors.New("must have unit in origin area")
		}
	}

	return nil
}

func validateSupport(order board.Order, from board.Area, to board.Area) error {
	if !from.HasNeighbor(order.To) {
		return errors.New("support order must be adjacent to destination")
	}

	return nil
}

func validateBesiegeOrTransport(order board.Order, from board.Area) error {
	if order.To != "" {
		return errors.New("besiege or transport orders cannot have destination")
	}

	switch order.Type {
	case board.OrderBesiege:
		return validateBesiege(order, from)
	case board.OrderTransport:
		return validateTransport(order, from)
	default:
		return errors.New("invalid order type")
	}
}

func validateBesiege(order board.Order, from board.Area) error {
	if !from.Castle {
		return errors.New("besieged area must have castle")
	}

	if from.IsControlled() {
		return errors.New("besieged area cannot already be controlled")
	}

	if from.Unit.Type == board.UnitShip {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(order board.Order, from board.Area) error {
	if from.Unit.Type != board.UnitShip {
		return errors.New("only ships can transport")
	}

	if !from.Sea {
		return errors.New("transport orders can only be placed at sea")
	}

	return nil
}

func validateWinterOrder(order board.Order, from board.Area, brd board.Board) error {
	switch order.Type {
	case board.OrderMove:
		return validateWinterMove(order, from, brd)
	case board.OrderBuild:
		return validateBuild(order, from, brd)
	}

	return errors.New("invalid order type")
}

func validateWinterMove(order board.Order, from board.Area, brd board.Board) error {
	if order.To == "" {
		return errors.New("winter move orders must have destination")
	}

	to, ok := brd.Areas[order.To]
	if !ok {
		return fmt.Errorf("destination area with name %s not found", order.To)
	}

	if to.ControllingPlayer != order.Player {
		return errors.New("must control destination area in winter move")
	}

	if from.Unit.Type == board.UnitShip && !to.IsCoast(brd) {
		return errors.New("ship winter move destination must be coast")
	}

	if order.Build != "" {
		return errors.New("cannot build unit with move order")
	}

	return nil
}

func validateBuild(order board.Order, from board.Area, brd board.Board) error {
	if !from.IsEmpty() {
		return errors.New("cannot build in area already occupied")
	}

	switch order.Build {
	case board.UnitShip:
		if !from.IsCoast(brd) {
			return errors.New("ships can only be built on coast")
		}
	case board.UnitFootman:
	case board.UnitHorse:
	case board.UnitCatapult:
	default:
		return errors.New("invalid unit type")
	}

	return nil
}

func validateOrderSet(orders []board.Order, brd board.Board) error {
	if err := validateUniqueMoveDestinations(orders, brd); err != nil {
		return err
	}

	if err := validateOneOrderPerArea(orders, brd); err != nil {
		return err
	}

	return nil
}

func validateUniqueMoveDestinations(orders []board.Order, brd board.Board) error {
	moveDestinations := make(map[string]struct{})

	for _, order := range orders {
		if order.Type == board.OrderMove {
			if _, notUnique := moveDestinations[order.To]; notUnique {
				return fmt.Errorf("orders include two moves to area %s", order.To)
			}

			moveDestinations[order.To] = struct{}{}
		}
	}

	return nil
}

func validateOneOrderPerArea(orders []board.Order, brd board.Board) error {
	orderedAreas := make(map[string]struct{})

	for _, order := range orders {
		if _, alreadyOrdered := orderedAreas[order.From]; alreadyOrdered {
			return fmt.Errorf("unit in area %s is ordered twice", order.From)
		}

		orderedAreas[order.From] = struct{}{}
	}

	return nil
}
