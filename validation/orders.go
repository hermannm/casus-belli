package validation

import (
	"errors"
	t "immerse-ntnu/hermannia/server/types"
	"immerse-ntnu/hermannia/server/utils"
)

func ValidateOrder(order t.Order) error {
	switch {
	case order.Type == t.Move || order.Type == t.Support:
		validateMoveOrSupport(order)
	case order.Type == t.Besiege || order.Type == t.Transport:
		validateBesiegeOrTransport(order)
	}

	return nil
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

	return nil
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

	return nil
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
