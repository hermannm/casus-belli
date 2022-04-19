package validation

import (
	"errors"

	"hermannm.dev/bfh-server/game/board"
)

func ValidateOrderSet(orders []board.Order, brd board.Board, season board.Season) error {
	for _, order := range orders {
		err := validateOrder(order, brd, season)
		if err != nil {
			return err
		}
	}

	return nil
}

// Takes a game order, and returns an error if it is invalid.
func validateOrder(order board.Order, brd board.Board, season board.Season) error {
	from := brd.Areas[order.From]

	if order.Player != from.Control {
		return errors.New("must control area that is ordered")
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
		return errors.New("units can only be built in winter")
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
		return errors.New("mvoes and supports must have destination")
	}

	to, ok := brd.Areas[order.To]
	if !ok {
		return errors.New("invalid order destination")
	}

	if !from.HasNeighbor(order.To) {
		return errors.New("destination not adjacent to origin")
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
		return validateMove(order, from, to)
	case board.OrderSupport:
		return validateSupport(order, from, to)
	}

	return errors.New("invalid order type")
}

func validateMove(order board.Order, from board.Area, to board.Area) error {
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
		return errors.New("invalid order destination")
	}

	if to.Control != order.Player {
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
		break
	case board.UnitHorse:
		break
	case board.UnitCatapult:
		break
	default:
		return errors.New("invalid unit type")
	}

	return nil
}
