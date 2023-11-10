package orderresolving

import (
	"sync"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/devlog/log"
)

type supportDeclaration struct {
	from gametypes.PlayerFaction
	to   gametypes.PlayerFaction // Blank if nobody were supported.
}

// Calls support from support orders to the given region, and appends modifiers to the given map.
func appendSupportModifiers(
	results map[gametypes.PlayerFaction]gametypes.Result,
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
		if support.to == "" {
			continue
		}

		result, isFaction := results[support.to]
		if isFaction {
			result.Parts = append(result.Parts, gametypes.SupportBonus(support.from))
			results[support.to] = result
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

	if includeDefender && !region.IsEmpty() && region.Unit.Faction == support.Faction {
		supportReceiver <- supportDeclaration{from: support.Faction, to: support.Faction}
		return
	}

	for _, move := range moves {
		if support.Faction == move.Faction {
			supportReceiver <- supportDeclaration{from: support.Faction, to: support.Faction}
			return
		}
	}

	battlers := make([]gametypes.PlayerFaction, 0, len(moves)+1)
	for _, move := range moves {
		battlers = append(battlers, move.Faction)
	}
	if includeDefender && !region.IsEmpty() {
		battlers = append(battlers, region.Unit.Faction)
	}

	if err := messenger.SendSupportRequest(
		support.Faction,
		support.Origin,
		region.Name,
		battlers,
	); err != nil {
		log.Error(err, "failed to send support request")
		supportReceiver <- supportDeclaration{from: support.Faction, to: ""}
		return
	}

	supported, err := messenger.AwaitSupport(support.Faction, support.Origin, region.Name)
	if err != nil {
		log.Errorf(err, "failed to receive support declaration from faction '%s'", support.Faction)
		supportReceiver <- supportDeclaration{from: support.Faction, to: ""}
		return
	}

	supportReceiver <- supportDeclaration{from: support.Faction, to: supported}
}
