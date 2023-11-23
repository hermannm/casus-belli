package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"hermannm.dev/enumnames"
	"hermannm.dev/set"
	"hermannm.dev/wrap"
)

// An order submitted by a player for one of their units in a given round.
type Order struct {
	Type    OrderType
	Faction PlayerFaction
	Origin  RegionName

	// For move and support orders: name of destination region.
	Destination RegionName

	// For move orders with horse units: optional name of second destination region to move to if
	// the first destination was reached.
	SecondDestination RegionName

	// For move orders: name of DangerZone the order tries to pass through, if any.
	ViaDangerZone DangerZone

	// For build orders: type of unit to build.
	Build UnitType

	// For move orders that lost a singleplayer battle or tied a multiplayer battle, and have to
	// fight their way back to their origin region. Must be false when submitting orders.
	Retreat bool

	// For move orders: the type of unit moved.
	// Server sets this field on the order, to keep track of units between battles. Since it's
	// private, it's excluded from messages to clients - they can deduce this from the Origin field.
	unitType UnitType
}

type OrderType uint8

const (
	// An order for a unit to move from one region to another.
	// Includes internal moves in winter.
	OrderMove OrderType = iota + 1

	// An order for a unit to support battles in adjacent regions.
	OrderSupport

	// For ship unit at sea: an order to transport a land unit across the sea.
	OrderTransport

	// For land unit in unconquered castle region: an order to besiege the castle.
	OrderBesiege

	// For player-controlled region in winter: an order for the type of unit to build in the region.
	OrderBuild
)

var orderNames = enumnames.NewMap(map[OrderType]string{
	OrderMove:      "Move",
	OrderSupport:   "Support",
	OrderTransport: "Transport",
	OrderBesiege:   "Besiege",
	OrderBuild:     "Build",
})

func (orderType OrderType) String() string {
	return orderNames.GetNameOrFallback(orderType, "INVALID")
}

func (order Order) isNone() bool {
	return order.Type == 0
}

func (order Order) unit() Unit {
	return Unit{Type: order.unitType, Faction: order.Faction}
}

// Checks if the order is a move of a horse unit with a second destination.
func (order Order) hasSecondHorseMove() bool {
	return order.Type == OrderMove && order.unitType == UnitHorse && order.SecondDestination != ""
}

// Returns the order with the original destination set as the origin, and the destination set as the
// original second destination. Assumes hasSecondHorseMove has already been called.
func (order Order) secondHorseMove() Order {
	order.Origin = order.Destination
	order.Destination = order.SecondDestination
	order.SecondDestination = ""
	return order
}

// Custom json.Marshaler implementation, to serialize uninitialized orders to null.
func (order Order) MarshalJSON() ([]byte, error) {
	if order.isNone() {
		return []byte("null"), nil
	}

	// Alias to avoid infinite loop of MarshalJSON.
	type orderAlias Order

	return json.Marshal(orderAlias(order))
}

func (order Order) logAttribute() slog.Attr {
	attributes := []any{
		slog.String("faction", string(order.Faction)),
		slog.String("origin", string(order.Origin)),
	}
	if order.Destination != "" {
		attributes = append(attributes, slog.String("destination", string(order.Destination)))
	}

	return slog.Group("order", attributes...)
}

func (game *Game) gatherAndValidateOrders() []Order {
	orderChans := make(map[PlayerFaction]chan []Order, len(game.PlayerFactions))
	for _, faction := range game.PlayerFactions {
		orderChan := make(chan []Order, 1)
		orderChans[faction] = orderChan
		go game.gatherAndValidateOrderSet(faction, orderChan)
	}

	var allOrders []Order
	factionOrders := make(map[PlayerFaction][]Order, len(orderChans))
	for faction, orderChan := range orderChans {
		orders := <-orderChan
		allOrders = append(allOrders, orders...)
		factionOrders[faction] = orders
	}

	if err := game.messenger.SendOrdersReceived(factionOrders); err != nil {
		game.log.Error(err)
	}

	return allOrders
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
func (game *Game) gatherAndValidateOrderSet(faction PlayerFaction, orderChan chan<- []Order) {
	for {
		if err := game.messenger.SendOrderRequest(faction, game.season); err != nil {
			game.log.Error(err)
			orderChan <- []Order{}
			return
		}

		orders, err := game.messenger.AwaitOrders(faction)
		if err != nil {
			game.log.Error(err)
			orderChan <- []Order{}
			return
		}

		for i, order := range orders {
			order.Faction = faction

			origin, ok := game.board[order.Origin]
			if ok && !origin.empty() && order.Type != OrderBuild {
				order.unitType = origin.Unit.Type
			}

			orders[i] = order
		}

		if err := validateOrders(orders, game.board, game.season); err != nil {
			game.log.Error(err)
			game.messenger.SendError(faction, err)
			continue
		}

		if err := game.messenger.SendOrdersConfirmation(faction); err != nil {
			game.log.Error(err)
		}

		orderChan <- orders
	}
}

// Checks if the given set of orders are valid for the state of the board in the given season.
// Assumes that all orders are from the same faction.
func validateOrders(orders []Order, board Board, season Season) error {
	var err error
	if season == SeasonWinter {
		err = validateWinterOrders(orders, board)
	} else {
		err = validateNonWinterOrders(orders, board)
	}
	return err
}

func validateWinterOrders(orders []Order, board Board) error {
	for _, order := range orders {
		origin, ok := board[order.Origin]
		if !ok {
			return wrap.Error(
				fmt.Errorf("origin region with name '%s' not found", order.Origin), "invalid order",
			)
		}

		if err := validateWinterOrder(order, origin, board); err != nil {
			return wrap.Errorf(err, "invalid winter order in region '%s'", order.Origin)
		}
	}

	if err := validateOrderSet(orders, board); err != nil {
		return wrap.Error(err, "invalid winter order set")
	}

	return nil
}

func validateWinterOrder(order Order, origin *Region, board Board) error {
	switch order.Type {
	case OrderMove:
		return validateWinterMove(order, origin, board)
	case OrderBuild:
		return validateBuild(order, origin, board)
	default:
		return fmt.Errorf("order type '%s' is invalid in winter", order.Type)
	}
}

func validateWinterMove(order Order, origin *Region, board Board) error {
	if order.Destination == "" {
		return errors.New("winter move orders must have destination")
	}

	to, ok := board[order.Destination]
	if !ok {
		return fmt.Errorf("destination region with name '%s' not found", order.Destination)
	}

	if to.ControllingFaction != order.Faction {
		return errors.New("must control destination region in winter move")
	}

	if origin.Unit.Type == UnitShip && !to.isCoast(board) {
		return errors.New("ship winter move destination must be coast")
	}

	if !order.Build.isNone() {
		return errors.New("cannot build unit with move order")
	}

	return nil
}

func validateBuild(order Order, origin *Region, board Board) error {
	if !origin.empty() {
		return errors.New("cannot build in region already occupied")
	}

	switch order.Build {
	case UnitShip:
		if !origin.isCoast(board) {
			return errors.New("ships can only be built on coast")
		}
	case UnitFootman:
	case UnitHorse:
	case UnitCatapult:
	default:
		return errors.New("invalid unit type")
	}

	return nil
}

func validateNonWinterOrders(orders []Order, board Board) error {
	for _, order := range orders {
		origin, ok := board[order.Origin]
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

func validateNonWinterOrder(order Order, origin *Region, board Board) error {
	if !order.Build.isNone() {
		return errors.New("build orders can only be placed in winter")
	}

	if order.Retreat {
		return errors.New("retreat orders can only be created by the server")
	}

	if order.Faction != origin.Unit.Faction {
		return errors.New("must have unit in ordered region")
	}

	switch order.Type {
	case OrderMove, OrderSupport:
		return validateMoveOrSupport(order, origin, board)
	case OrderBesiege, OrderTransport:
		return validateBesiegeOrTransport(order, origin)
	default:
		return fmt.Errorf("invalid order type '%s'", order.Type)
	}
}

func validateMoveOrSupport(order Order, origin *Region, board Board) error {
	if order.Destination == "" {
		return errors.New("moves and supports must have destination")
	}

	destination, ok := board[order.Destination]
	if !ok {
		return fmt.Errorf("destination region with name '%s' not found", order.Destination)
	}

	if origin.Unit.Type == UnitShip {
		if !(destination.Sea || destination.isCoast(board)) {
			return errors.New("ship order destination must be sea or coast")
		}
	} else {
		if destination.Sea {
			return errors.New("only ships can order to seas")
		}
	}

	switch order.Type {
	case OrderMove:
		return validateMove(order, origin, board)
	case OrderSupport:
		return validateSupport(order, origin, destination)
	}

	return errors.New("invalid order type")
}

func validateMove(order Order, origin *Region, board Board) error {
	if order.SecondDestination != "" {
		if origin.Unit.Type != UnitHorse {
			return errors.New(
				"second destinations for move orders can only be applied to horse units",
			)
		}

		if _, ok := board[order.SecondDestination]; !ok {
			return fmt.Errorf(
				"second destination region with name '%s' not found", order.SecondDestination,
			)
		}
	}

	return nil
}

func validateSupport(order Order, origin *Region, destination *Region) error {
	if !origin.hasNeighbor(order.Destination) {
		return errors.New("support order must be adjacent to destination")
	}

	return nil
}

func validateBesiegeOrTransport(order Order, origin *Region) error {
	if order.Destination != "" {
		return errors.New("besiege or transport orders cannot have destination")
	}

	switch order.Type {
	case OrderBesiege:
		return validateBesiege(order, origin)
	case OrderTransport:
		return validateTransport(order, origin)
	default:
		return errors.New("invalid order type")
	}
}

func validateBesiege(order Order, origin *Region) error {
	if !origin.Castle {
		return errors.New("besieged region must have castle")
	}

	if origin.controlled() {
		return errors.New("besieged region cannot already be controlled")
	}

	if origin.Unit.Type == UnitShip {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(order Order, origin *Region) error {
	if origin.Unit.Type != UnitShip {
		return errors.New("only ships can transport")
	}

	if !origin.Sea {
		return errors.New("transport orders can only be placed at sea")
	}

	return nil
}

func validateReachableMoveDestinations(orders []Order, board Board) error {
	boardCopy := make(Board, len(board))
	for regionName, region := range board {
		regionCopy := *region
		boardCopy[regionName] = &regionCopy
	}

	boardCopy.placeOrders(orders)

	for _, order := range orders {
		if order.Type != OrderMove {
			continue
		}

		if err := validateReachableMoveDestination(order, boardCopy); err != nil {
			return wrap.Errorf(
				err, "invalid move from '%s' to '%s'", order.Origin, order.Destination,
			)
		}

		if order.hasSecondHorseMove() {
			if err := validateReachableMoveDestination(
				order.secondHorseMove(),
				boardCopy,
			); err != nil {
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

func validateReachableMoveDestination(move Order, board Board) error {
	origin := board[move.Origin]

	if !origin.hasNeighbor(move.Destination) {
		canTransport, _, _ := board.findTransportPath(move.Origin, move.Destination)

		if !canTransport {
			return errors.New("regions not adjacent, and no transport path available")
		}
	}

	return nil
}

func validateOrderSet(orders []Order, board Board) error {
	if err := validateUniqueMoveDestinations(orders, board); err != nil {
		return err
	}

	if err := validateOneOrderPerRegion(orders, board); err != nil {
		return err
	}

	return nil
}

func validateUniqueMoveDestinations(orders []Order, board Board) error {
	moveDestinations := set.ArraySetWithCapacity[RegionName](len(orders))

	for _, order := range orders {
		if order.Type == OrderMove {
			if moveDestinations.Contains(order.Destination) {
				return fmt.Errorf("orders include two moves to region '%s'", order.Destination)
			}

			if order.SecondDestination != "" && moveDestinations.Contains(order.SecondDestination) {
				return fmt.Errorf(
					"orders include two moves to region '%s'", order.SecondDestination,
				)
			}

			moveDestinations.Add(order.Destination)
		}
	}

	return nil
}

func validateOneOrderPerRegion(orders []Order, board Board) error {
	orderedRegions := set.ArraySetWithCapacity[RegionName](len(orders))

	for _, order := range orders {
		if orderedRegions.Contains(order.Origin) {
			return fmt.Errorf("unit in region '%s' is ordered twice", order.Origin)
		}

		orderedRegions.Add(order.Origin)
	}

	return nil
}
