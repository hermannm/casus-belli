package validation

import (
	"errors"
	t "immerse-ntnu/hermannia/server/types"
	"immerse-ntnu/hermannia/server/utils"
)

func ValidateOrder(order t.Order, season t.Season) error {
	if order.Player.Color != order.From.Control {
		return errors.New("must control area that is ordered")
	}

	switch season {
	case t.Winter:
		return validateWinterOrder(order)
	default:
		return validateNonWinterOrder(order)
	}
}

func validateNonWinterOrder(order t.Order) error {
	if order.UnitBuild != "" {
		return errors.New("units can only be built in winter")
	}

	switch {
	case order.Type == t.Move || order.Type == t.Support:
		return validateMoveOrSupport(order)
	case order.Type == t.Besiege || order.Type == t.Transport:
		return validateBesiegeOrTransport(order)
	}

	return errors.New("invalid order type")
}

func validateMoveOrSupport(order t.Order) error {
	if order.To == nil {
		return errors.New("moves and supports must have destination")
	}

	if _, ok := order.From.Neighbors[order.To.Name]; !ok {
		return errors.New("destination not adjacent to origin")
	}

	if order.From.Unit.Type == t.Ship {
		if !utils.Sailable(*order.To) {
			return errors.New("ship order destination must be coast or sea")
		}
	} else {
		if order.To.Sea {
			return errors.New("only ships can order to seas")
		}
	}

	switch order.Type {
	case t.Move:
		return validateMove(order)
	case t.Support:
		return validateSupport(order)
	}

	return errors.New("invalid order type")
}

func validateMove(order t.Order) error {
	if order.SecondTo != nil && order.From.Unit.Type != t.Horse {
		return errors.New("only horse units can have second destination")
	}

	if order.From.Unit.Type == t.Horse {
		secondDestination, ok := order.To.Neighbors[order.SecondTo.Name]

		if secondDestination != nil && !ok {
			return errors.New("destination not adjacent to origin")
		}
	} else {
		if order.SecondTo != nil {
			return errors.New("only horse units can have second destination")
		}
	}

	return nil
}

func validateSupport(order t.Order) error {
	if order.SecondTo != nil {
		return errors.New("support orders cannot have second destination")
	}

	return nil
}

func validateBesiegeOrTransport(order t.Order) error {
	if order.To != nil || order.SecondTo != nil {
		return errors.New("besiege or transport orders cannot have destinations")
	}

	switch order.Type {
	case t.Besiege:
		return validateBesiege(order)
	case t.Transport:
		return validateTransport(order)
	}

	return errors.New("invalid order type")
}

func validateBesiege(order t.Order) error {
	if !order.From.Castle {
		return errors.New("besieged area must have castle")
	}

	if order.From.Control != t.Uncontrolled {
		return errors.New("besieged area cannot already be controlled")
	}

	if order.From.Unit.Type == t.Ship {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(order t.Order) error {
	if order.From.Unit.Type != t.Ship {
		return errors.New("only ships can transport")
	}

	return nil
}

func validateWinterOrder(order t.Order) error {
	if order.SecondTo != nil {
		return errors.New("winter order cannot have second destination")
	}

	switch order.Type {
	case t.Move:
		return validateWinterMove(order)
	case t.Build:
		return validateBuild(order)
	}

	return errors.New("invalid order type")
}

func validateWinterMove(order t.Order) error {
	if order.To.Control != order.Player.Color {
		return errors.New("must control destination area in winter move")
	}

	if order.UnitBuild != "" {
		return errors.New("cannot build unit with move order")
	}

	return nil
}

func validateBuild(order t.Order) error {
	if order.From.Unit != nil {
		return errors.New("cannot build in area already occupied")
	}

	switch order.UnitBuild {
	case t.Footman:
		break
	case t.Horse:
		break
	case t.Ship:
		break
	case t.Catapult:
		break
	default:
		return errors.New("invalid unit type")
	}

	return nil
}
