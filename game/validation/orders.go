package validation

import (
	"errors"
	"fmt"

	"hermannm.dev/bfh-server/game/gameboard"
)

// Checks if the given set of orders are valid for the state of the board in the given season.
// Assumes that all orders are from the same player.
func ValidateOrders(orders []gameboard.Order, board gameboard.Board, season gameboard.Season) error {
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

func validateOrder(order gameboard.Order, board gameboard.Board, season gameboard.Season) error {
	from := board.Regions[order.From]

	if order.Player != from.Unit.Player {
		return errors.New("must have unit in ordered region")
	}

	switch season {
	case gameboard.SeasonWinter:
		return validateWinterOrder(order, from, board)
	default:
		return validateNonWinterOrder(order, from, board)
	}
}

func validateNonWinterOrder(order gameboard.Order, from gameboard.Region, board gameboard.Board) error {
	if order.Build != "" {
		return errors.New("build orders can only be placed in winter")
	}

	switch order.Type {
	case gameboard.OrderMove:
		fallthrough
	case gameboard.OrderSupport:
		return validateMoveOrSupport(order, from, board)
	case gameboard.OrderBesiege:
		fallthrough
	case gameboard.OrderTransport:
		return validateBesiegeOrTransport(order, from)
	default:
		return errors.New("invalid order type")
	}
}

func validateMoveOrSupport(order gameboard.Order, from gameboard.Region, board gameboard.Board) error {
	if order.To == "" {
		return errors.New("moves and supports must have destination")
	}

	to, ok := board.Regions[order.To]
	if !ok {
		return fmt.Errorf("destination region with name %s not found", order.To)
	}

	if from.Unit.Type == gameboard.UnitShip {
		if !(to.Sea || to.IsCoast(board)) {
			return errors.New("ship order destination must be sea or coast")
		}
	} else {
		if to.Sea {
			return errors.New("only ships can order to seas")
		}
	}

	switch order.Type {
	case gameboard.OrderMove:
		return validateMove(order, from, to, board)
	case gameboard.OrderSupport:
		return validateSupport(order, from, to)
	}

	return errors.New("invalid order type")
}

func validateMove(order gameboard.Order, from gameboard.Region, to gameboard.Region, board gameboard.Board) error {
	if !from.HasNeighbor(order.To) {
		canTransport, _, _ := from.CanTransportTo(order.To, board)
		if !canTransport {
			return errors.New("move is not adjacent to destination, and cannot be transported")
		}
	}

	if from.IsEmpty() || from.Unit.Player != order.Player {
		secondHorseMove := false

		for _, firstOrder := range from.IncomingMoves {
			if from.Unit.Type == gameboard.UnitHorse &&
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

func validateSupport(order gameboard.Order, from gameboard.Region, to gameboard.Region) error {
	if !from.HasNeighbor(order.To) {
		return errors.New("support order must be adjacent to destination")
	}

	return nil
}

func validateBesiegeOrTransport(order gameboard.Order, from gameboard.Region) error {
	if order.To != "" {
		return errors.New("besiege or transport orders cannot have destination")
	}

	switch order.Type {
	case gameboard.OrderBesiege:
		return validateBesiege(order, from)
	case gameboard.OrderTransport:
		return validateTransport(order, from)
	default:
		return errors.New("invalid order type")
	}
}

func validateBesiege(order gameboard.Order, from gameboard.Region) error {
	if !from.Castle {
		return errors.New("besieged region must have castle")
	}

	if from.IsControlled() {
		return errors.New("besieged region cannot already be controlled")
	}

	if from.Unit.Type == gameboard.UnitShip {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(order gameboard.Order, from gameboard.Region) error {
	if from.Unit.Type != gameboard.UnitShip {
		return errors.New("only ships can transport")
	}

	if !from.Sea {
		return errors.New("transport orders can only be placed at sea")
	}

	return nil
}

func validateWinterOrder(order gameboard.Order, from gameboard.Region, board gameboard.Board) error {
	switch order.Type {
	case gameboard.OrderMove:
		return validateWinterMove(order, from, board)
	case gameboard.OrderBuild:
		return validateBuild(order, from, board)
	}

	return errors.New("invalid order type")
}

func validateWinterMove(order gameboard.Order, from gameboard.Region, board gameboard.Board) error {
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

	if from.Unit.Type == gameboard.UnitShip && !to.IsCoast(board) {
		return errors.New("ship winter move destination must be coast")
	}

	if order.Build != "" {
		return errors.New("cannot build unit with move order")
	}

	return nil
}

func validateBuild(order gameboard.Order, from gameboard.Region, board gameboard.Board) error {
	if !from.IsEmpty() {
		return errors.New("cannot build in region already occupied")
	}

	switch order.Build {
	case gameboard.UnitShip:
		if !from.IsCoast(board) {
			return errors.New("ships can only be built on coast")
		}
	case gameboard.UnitFootman:
	case gameboard.UnitHorse:
	case gameboard.UnitCatapult:
	default:
		return errors.New("invalid unit type")
	}

	return nil
}

func validateOrderSet(orders []gameboard.Order, board gameboard.Board) error {
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

func validateUniqueMoveDestinations(orders []gameboard.Order, board gameboard.Board) error {
	moveDestinations := make(map[string]struct{})

	for _, order := range orders {
		if order.Type == gameboard.OrderMove {
			_, notUnique := moveDestinations[order.To]
			if notUnique {
				return fmt.Errorf("orders include two moves to region %s", order.To)
			}

			moveDestinations[order.To] = struct{}{}
		}
	}

	return nil
}

func validateOneOrderPerRegion(orders []gameboard.Order, board gameboard.Board) error {
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
