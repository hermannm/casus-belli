package orderresolving

import (
	"fmt"
	"log"
	"sync"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/wrap"
)

type supportDeclaration struct {
	fromPlayer string
	toPlayer   string // Blank if supporting nobody.
}

// Calls support from support orders to the given region, and appends modifiers to the given map.
func appendSupportModifiers(
	results map[string]gametypes.Result,
	region gametypes.Region,
	includeDefender bool,
	messenger Messenger,
) {
	supports := region.IncomingSupports
	supportCount := len(supports)
	supportReceiver := make(chan supportDeclaration, supportCount)

	var waitGroup sync.WaitGroup
	waitGroup.Add(supportCount)

	incomingMoves := []gametypes.Order{}
	for _, result := range results {
		if result.DefenderRegion != "" {
			continue
		}
		if result.Move.Destination == region.Name {
			incomingMoves = append(incomingMoves, result.Move)
		}
	}

	for _, support := range supports {
		go callSupport(
			support,
			region,
			incomingMoves,
			includeDefender,
			supportReceiver,
			&waitGroup,
			messenger,
		)
	}

	waitGroup.Wait()
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

func callSupport(
	support gametypes.Order,
	region gametypes.Region,
	moves []gametypes.Order,
	includeDefender bool,
	supportReceiver chan<- supportDeclaration,
	waitGroup *sync.WaitGroup,
	messenger Messenger,
) {
	defer waitGroup.Done()

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

	if err := messenger.SendSupportRequest(
		support.Player,
		support.Origin,
		region.Name,
		battlers,
	); err != nil {
		fmt.Println(wrap.Error(err, "failed to send support request"))
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: ""}
		return
	}

	supported, err := messenger.AwaitSupport(support.Player, support.Origin, region.Name)
	if err != nil {
		log.Println(
			wrap.Errorf(
				err,
				"failed to receive support declaration from player '%s'",
				support.Player,
			),
		)
		supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: ""}
		return
	}

	supportReceiver <- supportDeclaration{fromPlayer: support.Player, toPlayer: supported}
}
