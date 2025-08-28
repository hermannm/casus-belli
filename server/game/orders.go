package game

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hermannm.dev/enumnames"
	"hermannm.dev/set"
	"hermannm.dev/wrap"
)

// An order submitted by a player for one of their units in a given round.
type Order struct {
	Type OrderType

	// For build orders: the type of unit moved.
	// For all other orders: the type of unit in the ordered region.
	UnitType UnitType

	// For move orders that lost a singleplayer battle or tied a multiplayer battle, and have to
	// fight their way back to their origin region. Must be false when submitting orders.
	Retreat bool

	// The faction of the player that submitted the order.
	Faction PlayerFaction

	// The region where the order was placed.
	Origin RegionName

	// For move and support orders: name of destination region.
	Destination RegionName

	// For move orders with knight units: optional name of second destination region to move to if
	// the first destination was reached.
	SecondDestination RegionName

	// For move orders: name of DangerZone the order tries to pass through, if any.
	ViaDangerZone DangerZone
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

	// For empty player-controlled region in winter: an order to build a unit in the region.
	OrderBuild

	// For region with a player's own unit, in winter: an order to disband the unit, when their
	// current number of units exceeds their max number of units.
	OrderDisband
)

var orderNames = enumnames.NewMap(
	map[OrderType]string{
		OrderMove:      "Move",
		OrderSupport:   "Support",
		OrderTransport: "Transport",
		OrderBesiege:   "Besiege",
		OrderBuild:     "Build",
	},
)

func (orderType OrderType) String() string {
	return orderNames.GetNameOrFallback(orderType, "INVALID")
}

func (order Order) unit() Unit {
	return Unit{Type: order.UnitType, Faction: order.Faction}
}

// Checks if the order is a move by a knight unit with a second destination.
func (order Order) hasKnightMove() bool {
	return order.Type == OrderMove && order.UnitType == UnitKnight && order.SecondDestination != ""
}

// Returns the order with the original destination set as the origin, and the destination set as the
// original second destination. Assumes hasKnightMove has already been called.
func (order Order) knightMove() Order {
	order.Origin = order.Destination
	order.Destination = order.SecondDestination
	order.SecondDestination = ""
	return order
}

func (order Order) mustCrossDangerZone(
	destination *Region,
) (mustCross bool, dangerZone DangerZone) {
	neighbor, adjacent := destination.getNeighbor(order.Origin, order.ViaDangerZone)
	if !adjacent {
		return false, "" // Non-adjacent moves are handled by transport resolving
	}

	return neighbor.DangerZone != "", neighbor.DangerZone
}

func countOrdersFromFaction(orders []Order, faction PlayerFaction) int {
	count := 0
	for _, order := range orders {
		if order.Faction == faction {
			count++
		}
	}
	return count
}

func (game *Game) gatherAndValidateOrders() []Order {
	ctx, cleanup := context.WithTimeoutCause(
		context.Background(),
		15*time.Minute,
		errors.New("timed out after 15 minutes"),
	)
	defer cleanup()

	orderChans := make(map[PlayerFaction]chan []Order, len(game.PlayerFactions))
	for _, faction := range game.PlayerFactions {
		orderChan := make(chan []Order, 1)
		orderChans[faction] = orderChan
		go game.gatherAndValidateOrderSet(ctx, faction, orderChan)
	}

	var allOrders []Order
	factionOrders := make(map[PlayerFaction][]Order, len(orderChans))
	for faction, orderChan := range orderChans {
		orders := <-orderChan
		allOrders = append(allOrders, orders...)
		factionOrders[faction] = orders
	}

	game.messenger.SendOrdersReceived(factionOrders)
	return allOrders
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
func (game *Game) gatherAndValidateOrderSet(
	ctx context.Context,
	faction PlayerFaction,
	orderChan chan<- []Order,
) {
	for {
		if succeeded := game.messenger.SendOrderRequest(faction, game.season); !succeeded {
			orderChan <- []Order{}
			return
		}

		orders, err := game.messenger.AwaitOrders(ctx, faction)
		if err != nil {
			err = wrap.Error(err, "failed to receive orders")
			game.log.Error(ctx, err, "")
			game.messenger.SendError(faction, err)
			orderChan <- []Order{}
			return
		}

		for i := range orders {
			orders[i].Faction = faction
		}

		if err := validateOrders(orders, faction, game.board, game.season); err != nil {
			game.log.Error(ctx, err, "")
			game.messenger.SendError(faction, err)
			continue
		}

		game.messenger.SendOrdersConfirmation(faction)
		orderChan <- orders
	}
}

// Checks if the given set of orders are valid for the state of the board in the given season.
// Assumes that all orders are from the same faction.
func validateOrders(orders []Order, faction PlayerFaction, board Board, season Season) error {
	var err error
	if season == SeasonWinter {
		err = validateWinterOrders(orders, faction, board)
	} else {
		err = validateNonWinterOrders(orders, board)
	}
	return err
}

func validateWinterOrders(orders []Order, faction PlayerFaction, board Board) error {
	var disbands set.ArraySet[RegionName]
	var outgoingMoves set.ArraySet[RegionName]
	for _, order := range orders {
		if order.Type == OrderDisband {
			disbands.Add(order.Origin)
		} else if order.Type == OrderMove {
			outgoingMoves.Add(order.Origin)
		}
	}

	for _, order := range orders {
		origin, ok := board[order.Origin]
		if !ok {
			return wrap.Error(
				fmt.Errorf("origin region with name '%s' not found", order.Origin), "invalid order",
			)
		}

		if err := validateWinterOrder(order, origin, board, disbands, outgoingMoves); err != nil {
			return wrap.Errorf(err, "invalid winter order in region '%s'", order.Origin)
		}
	}

	if err := validateOrderSet(orders); err != nil {
		return wrap.Error(err, "invalid winter order set")
	}

	if err := validateNumberOfBuilds(orders, faction, board, disbands); err != nil {
		return wrap.Error(err, "invalid winter order set")
	}

	return nil
}

func validateWinterOrder(
	order Order,
	origin *Region,
	board Board,
	disbands set.ArraySet[RegionName],
	outgoingMoves set.ArraySet[RegionName],
) error {
	if err := validateOrderedUnit(order, origin); err != nil {
		return err
	}

	switch order.Type {
	case OrderMove:
		return validateWinterMove(order, origin, board, disbands, outgoingMoves)
	case OrderBuild:
		return validateBuild(order, origin, board)
	case OrderDisband:
		// No extra validation needed - validateOrderedUnit already checks that the ordered region
		// is not empty, and that its unit matches the submitting player's faction
		return nil
	default:
		return fmt.Errorf("order type '%s' is invalid in winter", order.Type)
	}
}

func validateWinterMove(
	order Order,
	origin *Region,
	board Board,
	disbands set.ArraySet[RegionName],
	outgoingMoves set.ArraySet[RegionName],
) error {
	if order.Destination == "" {
		return errors.New("winter move orders must have destination")
	}

	destination, ok := board[order.Destination]
	if !ok {
		return fmt.Errorf("destination region with name '%s' not found", order.Destination)
	}

	if destination.ControllingFaction != order.Faction {
		return errors.New("must control destination region in winter move")
	}

	if !destination.empty() &&
		!disbands.Contains(destination.Name) &&
		!outgoingMoves.Contains(destination.Name) {
		return fmt.Errorf("move destination '%s' already has a unit", destination.Name)
	}

	if origin.Unit.Value.Type == UnitShip && !destination.isCoast(board) {
		return errors.New("ship winter move destination must be coast")
	}

	return nil
}

func validateBuild(order Order, origin *Region, board Board) error {
	if !origin.empty() {
		return errors.New("cannot build in region already occupied")
	}

	switch order.UnitType {
	case UnitShip:
		if !origin.isCoast(board) {
			return errors.New("ships can only be built on coast")
		}
	case UnitFootman, UnitKnight, UnitCatapult:
		// Valid
	default:
		return errors.New("invalid unit type")
	}

	return nil
}

func validateNumberOfBuilds(
	orders []Order,
	faction PlayerFaction,
	board Board,
	disbands set.ArraySet[RegionName],
) error {
	unitCount, maxUnitCount := board.unitCounts(faction)
	unitsToBuild := maxUnitCount - unitCount

	buildOrderCount := 0
	for _, order := range orders {
		if order.Type == OrderBuild {
			buildOrderCount++
		}
	}

	if unitsToBuild < 0 {
		unitsToDisband := -unitsToBuild
		if buildOrderCount != 0 {
			return fmt.Errorf(
				"cannot place build orders when you need to disband units (%d units to disband)",
				unitsToDisband,
			)
		}
		if disbands.Size() != unitsToDisband {
			return fmt.Errorf(
				"need to disband %d units, but received %d disband orders",
				unitsToDisband,
				disbands.Size(),
			)
		}
		return nil
	}

	if buildOrderCount > unitsToBuild {
		return fmt.Errorf(
			"have %d units to build, but received %d build orders",
			unitsToBuild,
			buildOrderCount,
		)
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

	if err := validateOrderSet(orders); err != nil {
		return wrap.Error(err, "invalid order set")
	}

	if err := validateReachableMoveDestinations(orders, board); err != nil {
		return err
	}

	return nil
}

func validateNonWinterOrder(order Order, origin *Region, board Board) error {
	if err := validateOrderedUnit(order, origin); err != nil {
		return err
	}

	if order.Retreat {
		return errors.New("retreat orders can only be created by the server")
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

	if origin.Unit.Value.Type == UnitShip {
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
		return validateSupport(order, origin)
	default:
		return errors.New("invalid order type")
	}
}

func validateMove(order Order, origin *Region, board Board) error {
	if order.SecondDestination != "" {
		if origin.Unit.Value.Type != UnitKnight {
			return errors.New(
				"second destinations for move orders can only be applied to knight units",
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

func validateSupport(order Order, origin *Region) error {
	if !origin.adjacentTo(order.Destination) {
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
		return validateBesiege(origin)
	case OrderTransport:
		return validateTransport(origin)
	default:
		return errors.New("invalid order type")
	}
}

func validateBesiege(origin *Region) error {
	if !origin.Castle {
		return errors.New("besieged region must have castle")
	}

	if origin.controlled() {
		return errors.New("besieged region cannot already be controlled")
	}

	if origin.Unit.Value.Type == UnitShip {
		return errors.New("ships cannot besiege")
	}

	return nil
}

func validateTransport(origin *Region) error {
	if origin.Unit.Value.Type != UnitShip {
		return errors.New("only ships can transport")
	}

	if !origin.Sea {
		return errors.New("transport orders can only be placed at sea")
	}

	return nil
}

func validateReachableMoveDestinations(orders []Order, board Board) error {
	// Copy the board, so orders do not persist
	board = board.copy()
	board.placeOrders(orders)

	for _, order := range orders {
		if order.Type != OrderMove {
			continue
		}

		if err := validateReachableMoveDestination(order, board); err != nil {
			return wrap.Errorf(
				err, "invalid move from '%s' to '%s'", order.Origin, order.Destination,
			)
		}

		if order.hasKnightMove() {
			if err := validateReachableMoveDestination(order.knightMove(), board); err != nil {
				return wrap.Errorf(
					err,
					"invalid second destination for knight move from '%s' to '%s'",
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

	if !origin.adjacentTo(move.Destination) {
		canTransport, _, _ := board.findTransportPath(move.Origin, move.Destination)

		if !canTransport {
			return errors.New("regions not adjacent, and no transport path available")
		}
	}

	return nil
}

func validateOrderSet(orders []Order) error {
	if err := validateUniqueMoveDestinations(orders); err != nil {
		return err
	}

	if err := validateOneOrderPerRegion(orders); err != nil {
		return err
	}

	return nil
}

func validateUniqueMoveDestinations(orders []Order) error {
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

func validateOneOrderPerRegion(orders []Order) error {
	orderedRegions := set.ArraySetWithCapacity[RegionName](len(orders))

	for _, order := range orders {
		if orderedRegions.Contains(order.Origin) {
			return fmt.Errorf("unit in region '%s' is ordered twice", order.Origin)
		}

		orderedRegions.Add(order.Origin)
	}

	return nil
}

func validateOrderedUnit(order Order, origin *Region) error {
	if !order.UnitType.isValid() {
		return fmt.Errorf("invalid ordered unit type '%d'", order.UnitType)
	}

	if order.Type != OrderBuild {
		if origin.empty() {
			return errors.New("ordered region does not have a unit")
		}

		if origin.Unit.Value.Faction != order.Faction {
			return fmt.Errorf(
				"faction of ordered unit '%s' does not match your faction '%s'",
				origin.Unit.Value.Faction,
				order.Faction,
			)
		}

		if origin.Unit.Value.Type != order.UnitType {
			return fmt.Errorf(
				"order unit type '%v' does not match unit type '%v' in ordered region",
				order.UnitType,
				origin.Unit.Value.Type,
			)
		}
	}

	return nil
}
