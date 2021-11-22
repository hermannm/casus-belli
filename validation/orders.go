package validation

import (
	"errors"
	. "immerse-ntnu/hermannia/server/types"
	"immerse-ntnu/hermannia/server/utils"
)

func ValidateOrder(order Order, season Season) error {
	if order.Player.Color != order.From.Control {
		return errors.New("must control area that is ordered")
	}

	switch season {
	case Winter:
		return validateWinterOrder(order)
	default:
		return validateNonWinterOrder(order)
	}
}

func validateNonWinterOrder(order Order) error {
	if order.UnitBuild != "" {
		return errors.New("units can only be built in winter")
	}

	switch {
	case order.Type == Move || order.Type == Support:
		return validateMoveOrSupport(order)
	case order.Type == Besiege || order.Type == Transport:
		return validateBesiegeOrTransport(order)
	}

	return errors.New("invalid order type")
}

func validateMoveOrSupport(order Order) error {
	if order.To == nil {
		return errors.New("moves and supports must have destination")
	}

	if _, ok := order.From.Neighbors[order.To.Name]; !ok {
		return errors.New("destination not adjacent to origin")
	}

	if order.From.Unit.Type == Ship {
		if !utils.Sailable(*order.To) {
			return errors.New("ship order destination must be coast or sea")
		}
	} else {
		if order.To.Sea {
			return errors.New("only ships can order to seas")
		}
	}

	switch order.Type {
	case Move:
		return validateMove(order)
	case Support:
		return validateSupport(order)
	}

	return errors.New("invalid order type")
}

func validateMove(order Order) error {
	if order.From.Unit == nil || order.From.Unit.Color != order.Player.Color {
		eligibleDep := false

		for _, dep := range order.Dependencies {
			if !(dep.From.Unit.Type == Horse &&
				dep.From.Unit.Color == order.Player.Color &&
				dep.To == order.From) {
				eligibleDep = true
			}
		}

		if !eligibleDep {
			return errors.New("must have unit in origin area")
		}
	}

	return nil
}

func validateSupport(order Order) error {
	return nil
}

func validateBesiegeOrTransport(order Order) error {
	if order.To != nil {
		return errors.New("besiege or transport orders cannot have destination")
	}

	switch order.Type {
	case Besiege:
		return validateBesiege(order)
	case Transport:
		return validateTransport(order)
	}

	return errors.New("invalid order type")
}

func validateBesiege(order Order) error {
	if !order.From.Castle {
		return errors.New("besieged area must have castle")
	}

	if order.From.Control != Uncontrolled {
		return errors.New("besieged area cannot already be controlled")
	}

	if order.From.Unit.Type == Ship {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(order Order) error {
	if order.From.Unit.Type != Ship {
		return errors.New("only ships can transport")
	}

	return nil
}

func validateWinterOrder(order Order) error {
	switch order.Type {
	case Move:
		return validateWinterMove(order)
	case Build:
		return validateBuild(order)
	}

	return errors.New("invalid order type")
}

func validateWinterMove(order Order) error {
	if order.To.Control != order.Player.Color {
		return errors.New("must control destination area in winter move")
	}

	if order.UnitBuild != "" {
		return errors.New("cannot build unit with move order")
	}

	return nil
}

func validateBuild(order Order) error {
	if order.From.Unit != nil {
		return errors.New("cannot build in area already occupied")
	}

	switch order.UnitBuild {
	case Footman:
		break
	case Horse:
		break
	case Ship:
		break
	case Catapult:
		break
	default:
		return errors.New("invalid unit type")
	}

	return nil
}
