package ordervalidation

import (
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/devlog/log"
)

type Messenger interface {
	AwaitOrders(from gametypes.PlayerFaction) ([]gametypes.Order, error)
	SendOrderRequest(to gametypes.PlayerFaction) error
	SendOrdersConfirmation(factionThatSubmittedOrders gametypes.PlayerFaction) error
	SendOrdersReceived(orders map[gametypes.PlayerFaction][]gametypes.Order) error
	SendError(to gametypes.PlayerFaction, err error)
}

func GatherAndValidateOrders(
	factions []gametypes.PlayerFaction,
	board gametypes.Board,
	season gametypes.Season,
	messenger Messenger,
) []gametypes.Order {
	orderChans := make(map[gametypes.PlayerFaction]chan []gametypes.Order, len(factions))
	for _, faction := range factions {
		orderChan := make(chan []gametypes.Order, 1)
		orderChans[faction] = orderChan
		go gatherAndValidateOrderSet(faction, board, season, orderChan, messenger)
	}

	var allOrders []gametypes.Order
	factionOrders := make(map[gametypes.PlayerFaction][]gametypes.Order, len(orderChans))
	for faction, orderChan := range orderChans {
		orders := <-orderChan
		allOrders = append(allOrders, orders...)
		factionOrders[faction] = orders
	}

	if err := messenger.SendOrdersReceived(factionOrders); err != nil {
		log.Error(err, "")
	}

	return allOrders
}

// Waits for the given player to submit orders, then validates them.
// If valid, sends the order set to the given output channel.
// If invalid, informs the client and waits for a new order set.
func gatherAndValidateOrderSet(
	faction gametypes.PlayerFaction,
	board gametypes.Board,
	season gametypes.Season,
	orderChan chan<- []gametypes.Order,
	messenger Messenger,
) {
	for {
		if err := messenger.SendOrderRequest(faction); err != nil {
			log.Error(err, "")
			orderChan <- []gametypes.Order{}
			return
		}

		orders, err := messenger.AwaitOrders(faction)
		if err != nil {
			log.Error(err, "")
			orderChan <- []gametypes.Order{}
			return
		}

		for i, order := range orders {
			order.Faction = faction

			origin, ok := board.Regions[order.Origin]
			if ok && !origin.IsEmpty() && order.Type != gametypes.OrderBuild {
				order.Unit = origin.Unit
			}

			orders[i] = order
		}

		if err := validateOrders(orders, board, season); err != nil {
			log.Error(err, "")
			messenger.SendError(faction, err)
			continue
		}

		if err := messenger.SendOrdersConfirmation(faction); err != nil {
			log.Error(err, "")
		}

		orderChan <- orders
	}
}

// Checks if the given set of orders are valid for the state of the board in the given season.
// Assumes that all orders are from the same faction.
func validateOrders(
	orders []gametypes.Order,
	board gametypes.Board,
	season gametypes.Season,
) error {
	var err error
	if season == gametypes.SeasonWinter {
		err = validateWinterOrders(orders, board)
	} else {
		err = validateNonWinterOrders(orders, board)
	}
	return err
}
