package game

import (
	"errors"
)

func validateOrderSet(orders []Order, board Board, season Season) error {
	for _, order := range orders {
		err := validateOrder(order, board, season)
		if err != nil {
			return err
		}
	}

	return nil
}

// Takes a game order, and returns an error if it is invalid.
func validateOrder(order Order, board Board, season Season) error {
	from := board[order.From]

	if order.Player != from.Control {
		return errors.New("must control area that is ordered")
	}

	switch season {
	case Winter:
		return validateWinterOrder(order, from, board)
	default:
		return validateNonWinterOrder(order, from, board)
	}
}

func validateNonWinterOrder(order Order, from Area, board Board) error {
	if order.Build != "" {
		return errors.New("units can only be built in winter")
	}

	switch order.Type {
	case Move:
		fallthrough
	case Support:
		return validateMoveOrSupport(order, from, board)
	case Besiege:
		fallthrough
	case Transport:
		return validateBesiegeOrTransport(order, from)
	default:
		return errors.New("invalid order type")
	}
}

func validateMoveOrSupport(order Order, from Area, board Board) error {
	if order.To == "" {
		return errors.New("mvoes and supports must have destination")
	}

	to, ok := board[order.To]
	if !ok {
		return errors.New("invalid order destination")
	}

	if !from.HasNeighbor(order.To) {
		return errors.New("destination not adjacent to origin")
	}

	if from.Unit.Type == Ship {
		if !(to.Sea || to.IsCoast(board)) {
			return errors.New("ship order destination must be sea or coast")
		}
	} else {
		if to.Sea {
			return errors.New("only ships can order to seas")
		}
	}

	switch order.Type {
	case Move:
		return validateMove(order, from, to)
	case Support:
		return validateSupport(order, from, to)
	}

	return errors.New("invalid order type")
}

func validateMove(order Order, from Area, to Area) error {
	if from.IsEmpty() || from.Unit.Player != order.Player {
		secondHorseMove := false

		for _, firstOrder := range from.IncomingMoves {
			if from.Unit.Type == Horse &&
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

func validateSupport(order Order, from Area, to Area) error {
	return nil
}

func validateBesiegeOrTransport(order Order, from Area) error {
	if order.To != "" {
		return errors.New("besiege or transport orders cannot have destination")
	}

	switch order.Type {
	case Besiege:
		return validateBesiege(order, from)
	case Transport:
		return validateTransport(order, from)
	default:
		return errors.New("invalid order type")
	}
}

func validateBesiege(order Order, from Area) error {
	if !from.Castle {
		return errors.New("besieged area must have castle")
	}

	if from.IsControlled() {
		return errors.New("besieged area cannot already be controlled")
	}

	if from.Unit.Type == Ship {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(order Order, from Area) error {
	if from.Unit.Type != Ship {
		return errors.New("only ships can transport")
	}

	return nil
}

func validateWinterOrder(order Order, from Area, board Board) error {
	switch order.Type {
	case Move:
		return validateWinterMove(order, from, board)
	case Build:
		return validateBuild(order, from, board)
	}

	return errors.New("invalid order type")
}

func validateWinterMove(order Order, from Area, board Board) error {
	if order.To == "" {
		return errors.New("winter move orders must have destination")
	}

	to, ok := board[order.To]
	if !ok {
		return errors.New("invalid order destination")
	}

	if to.Control != order.Player {
		return errors.New("must control destination area in winter move")
	}

	if from.Unit.Type == Ship && !to.IsCoast(board) {
		return errors.New("ship winter move destination must be coast")
	}

	if order.Build != "" {
		return errors.New("cannot build unit with move order")
	}

	return nil
}

func validateBuild(order Order, from Area, board Board) error {
	if !from.IsEmpty() {
		return errors.New("cannot build in area already occupied")
	}

	switch order.Build {
	case Ship:
		if !from.IsCoast(board) {
			return errors.New("ships can only be built on coast")
		}
	case Footman:
		break
	case Horse:
		break
	case Catapult:
		break
	default:
		return errors.New("invalid unit type")
	}

	return nil
}
