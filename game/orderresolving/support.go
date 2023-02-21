package orderresolving

import (
	"fmt"
	"log"
	"sync"

	"hermannm.dev/bfh-server/game/gametypes"
)

type supportDeclaration struct {
	fromPlayer string
	toPlayer   string
}

// Calls support from support orders to the given region.
// Appends support modifiers to receiving players' results in the given map,
// but only if the result is tied to a move order to the region.
// Calls support to defender in the region if includeDefender is true.
func appendSupportMods(
	results map[string]gametypes.Result,
	region gametypes.Region,
	includeDefender bool,
	messenger Messenger,
) {
	supports := region.IncomingSupports
	supportCount := len(supports)
	supportReceiver := make(chan supportDeclaration, supportCount)
	var wg sync.WaitGroup
	wg.Add(supportCount)

	// Finds the moves going to this region.
	moves := []gametypes.Order{}
	for _, result := range results {
		if result.DefenderRegion != "" {
			continue
		}
		if result.Move.Destination == region.Name {
			moves = append(moves, result.Move)
		}
	}

	// Starts a goroutine to call support for each support order to the region.
	for _, support := range supports {
		go callSupport(support, region, moves, includeDefender, supportReceiver, &wg, messenger)
	}

	// Waits until all support calls are done, then closes the channel to range over it.
	wg.Wait()
	close(supportReceiver)

	for support := range supportReceiver {
		if support.toPlayer == "" {
			continue
		}

		result, isPlayer := results[support.toPlayer]
		if isPlayer {
			result.Parts = append(result.Parts, gametypes.SupportBonus(support.fromPlayer))
			results[support.toPlayer] = result
		}
	}
}

// Finds out which player a given support order supports in a battle. Sends the resulting support
// declaration to the given supportReceiver, and decrements the wait group by 1.
//
// If the support order's player matches a player in the battle, support is automatically given to
// themselves.
// If support is not given to any player in the battle, the to field on the declaration is "".
func callSupport(
	support gametypes.Order,
	region gametypes.Region,
	moves []gametypes.Order,
	includeDefender bool,
	supportReceiver chan<- supportDeclaration,
	wg *sync.WaitGroup,
	messenger Messenger,
) {
	defer wg.Done()

	if includeDefender && !region.IsEmpty() && region.Unit.Player == support.Player {
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: support.Player}
		return
	}

	for _, move := range moves {
		if support.Player == move.Player {
			supportReceiver <- supportDeclaration{
				fromPlayer: support.Player,
				toPlayer:   support.Player,
			}
			return
		}
	}

	var battlers []string
	for _, move := range moves {
		battlers = append(battlers, move.Player)
	}
	if includeDefender && !region.IsEmpty() {
		battlers = append(battlers, region.Unit.Player)
	}

	err := messenger.SendSupportRequest(support.Player, region.Name, battlers)
	if err != nil {
		log.Println(fmt.Errorf("failed to send support request: %w", err))
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: ""}
		return
	}

	supported, err := messenger.ReceiveSupport(support.Player, region.Name)
	if err != nil {
		log.Println(fmt.Errorf(
			"failed to receive support declaration from player %s: %w",
			support.Player,
			err,
		))
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: ""}
		return
	}

	supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: supported}
}
