package validation

import (
	"errors"
	"immerse/hermannia/server/types"
	"immerse/hermannia/server/utils"
)

func ValidateOrder(order types.Order) error {
	switch {
	case order.Type == types.Move || order.Type == types.Support:
		validateMoveOrSupport(order)
	case order.Type == types.Besiege || order.Type == types.Transport:
		validateBesiegeOrTransport(order)
	}

	return nil
}

func validateMoveOrSupport(order types.Order) error {
	if _, ok := order.From.Neighbors[order.To.Name]; !ok {
		return errors.New("destination not adjacent to origin")
	}

	if order.From.OccupyingUnit.Type == types.Ship {
		if !utils.Sailable(*order.To) {
			return errors.New("ship order destination must be coast or sea")
		}
	} else {
		if order.To.Sea {
			return errors.New("only ships can order to seas")
		}
	}

	switch order.Type {
	case types.Move:
		return validateMove(order)
	case types.Support:
		return validateSupport(order)
	}

	return nil
}

func validateMove(order types.Order) error {
	if order.SecondTo != nil && order.From.OccupyingUnit.Type != types.Horse {
		return errors.New("only horse units can have second destination")
	}

	if order.From.OccupyingUnit.Type == types.Horse {
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

func validateSupport(order types.Order) error {
	if order.SecondTo != nil {
		return errors.New("support orders cannot have second destination")
	}

	return nil
}

func validateBesiegeOrTransport(order types.Order) error {
	if order.To != nil || order.SecondTo != nil {
		return errors.New("besiege or transport orders cannot have destinations")
	}

	switch order.Type {
	case types.Besiege:
		return validateBesiege(order)
	case types.Transport:
		return validateTransport(order)
	}

	return nil
}

func validateBesiege(order types.Order) error {
	if !order.From.Castle {
		return errors.New("besieged area must have castle")
	}

	if order.From.Control != "" {
		return errors.New("besieged area cannot already be conquered")
	}

	if order.From.OccupyingUnit.Type == types.Ship {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(order types.Order) error {
	if order.From.OccupyingUnit.Type != types.Ship {
		return errors.New("only ships can transport")
	}

	return nil
}
