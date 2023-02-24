package orderresolving

import (
	"log"

	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/set"
)

// Resolves moves on the board. Returns any resulting battles.
// Only resolves battles between players if allowPlayerConflict is true.
func resolveMoves(
	board gametypes.Board, allowPlayerConflict bool, messenger Messenger,
) []gametypes.Battle {
	var battles []gametypes.Battle

	battleReceiver := make(chan gametypes.Battle)
	processing := set.New[string]()
	processed := set.New[string]()
	retreats := make(map[string]gametypes.Order)

OuterLoop:
	for {
		select {
		case battle := <-battleReceiver:
			battles = append(battles, battle)
			messenger.SendBattleResults([]gametypes.Battle{battle})

			newRetreats := resolveBattle(battle, board)
			for _, retreat := range newRetreats {
				retreats[retreat.Origin] = retreat
			}

			for _, region := range battle.RegionNames() {
				processing.Remove(region)
			}
		default:
		BoardLoop:
			for regionName, region := range board.Regions {
				retreat, hasRetreat := retreats[regionName]

				if processed.Contains(regionName) && !hasRetreat {
					continue BoardLoop
				}

				if processing.Contains(regionName) {
					continue BoardLoop
				}

				for _, move := range region.IncomingMoves {
					transportAttacked, dangerZones := resolveTransports(move, region, board)

					if transportAttacked {
						if allowPlayerConflict {
							continue BoardLoop
						} else {
							processed.Add(regionName)
						}
					} else if len(dangerZones) > 0 {
						survived, dangerZoneCrossings := crossDangerZones(move, dangerZones)
						if !survived {
							board.RemoveOrder(move)
						}

						battles = append(battles, dangerZoneCrossings...)
						err := messenger.SendBattleResults(dangerZoneCrossings)
						if err != nil {
							log.Println(err)
						}
					}
				}

				if !region.IsAttacked() {
					if hasRetreat && region.IsEmpty() {
						region.Unit = retreat.Unit
						board.Regions[regionName] = region
						delete(retreats, regionName)
					}

					processed.Add(region.Name)
					continue BoardLoop
				}

				resolveRegionMoves(
					region,
					board,
					allowPlayerConflict,
					battleReceiver,
					processing,
					processed,
					messenger,
				)
			}

			if processing.IsEmpty() && len(retreats) == 0 {
				break OuterLoop
			}
		}
	}

	return battles
}

// Resolves moves to the given region on the board.
// Assumes that the region has incoming moves.
//
// Immediately resolves regions that do not require battle, and adds them to the given processed
// set.
//
// Adds embattled regions to the given processing set, and forwards them to appropriate battle
// calculation functions, which send results to the given battleReceiver.
//
// Skips regions that have outgoing moves, unless they are part of a move cycle.
// If allowPlayerConflict is false, skips regions that require battle between players.
func resolveRegionMoves(
	region gametypes.Region,
	board gametypes.Board,
	allowPlayerConflict bool,
	battleReceiver chan gametypes.Battle,
	processing set.Set[string],
	processed set.Set[string],
	messenger Messenger,
) {
	// Finds out if the move is part of a two-way cycle (moves moving against each other),
	// and resolves it.
	twoWayCycle, region2, samePlayer := discoverTwoWayCycle(region, board)
	if twoWayCycle {
		if samePlayer {
			// If both moves are by the same player, removes the units from their origin regions,
			// as they may not be allowed to retreat if their origin region is taken.
			for _, cycleRegion := range [2]gametypes.Region{region, region2} {
				cycleRegion.Unit = gametypes.Unit{}
				cycleRegion.Order = gametypes.Order{}
				board.Regions[cycleRegion.Name] = cycleRegion
			}
		} else {
			// If the moves are from different players, they battle in the middle.
			go calculateBorderBattle(region, region2, battleReceiver, messenger)
			processing.Add(region.Name, region2.Name)
			return
		}
	} else {
		// If there is a cycle longer than 2 moves, forwards the resolving to 'resolveCycle'.
		cycle, _ := discoverCycle(region.Name, region.Order, board)
		if cycle != nil {
			resolveCycle(
				cycle,
				board,
				allowPlayerConflict,
				battleReceiver,
				processing,
				processed,
				messenger,
			)
			return
		}
	}

	// Empty regions with only a single incoming move are either auto-successes or a singleplayer
	// battle.
	if len(region.IncomingMoves) == 1 && region.IsEmpty() {
		move := region.IncomingMoves[0]

		if region.IsControlled() || region.Sea {
			succeedMove(move, board)
			processed.Add(region.Name)
			return
		}

		go calculateSingleplayerBattle(region, move, battleReceiver, messenger)
		processing.Add(region.Name)
		return
	}

	// If the destination region has an outgoing move order, that must be resolved first.
	if region.Order.Type == gametypes.OrderMove {
		return
	}

	// If the function has not returned yet, then it must be a multiplayer battle.
	go calculateMultiplayerBattle(region, !region.IsEmpty(), battleReceiver, messenger)
	processing.Add(region.Name)
}
