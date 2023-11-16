package game_test

import (
	"testing"

	"hermannm.dev/bfh-server/game"
	"hermannm.dev/devlog/log"
)

func newMockGame() *game.Game {
	board := make(game.Board)

	regions := []game.Region{
		{Name: "Lusía", HasCastle: true},
		{Name: "Lomone", IsForest: true},
		{Name: "Limbol", IsForest: true},
		{Name: "Leil"},
		{Name: "Worp", IsForest: true, HomeFaction: "green", ControllingFaction: "green"},
		{
			Name:               "Winde",
			IsForest:           true,
			HasCastle:          true,
			HomeFaction:        "green",
			ControllingFaction: "green",
		},
		{Name: "Ovo", IsForest: true},
		{Name: "Mare Gond", IsSea: true},
		{Name: "Mare Elle", IsSea: true},
		{Name: "Zona"},
		{Name: "Tond"},
		{Name: "Tige"},
		{Name: "Tusser"},
		{Name: "Mare Ovond", IsSea: true},
		{Name: "Furie", HasCastle: true},
		{Name: "Firril"},
		{Name: "Fond"},
		{Name: "Gron"},
		{Name: "Gnade"},
		{Name: "Gewel", IsForest: true, HasCastle: true},
		{Name: "Mare Unna", IsSea: true},
		{Name: "Emman", IsForest: true, HomeFaction: "black", ControllingFaction: "black"},
		{Name: "Erren", HasCastle: true, HomeFaction: "black", ControllingFaction: "black"},
		{Name: "Mare Bøso", IsSea: true},
	}

	// Defines a utility struct for two-way neighbor declaration, to avoid repetition.
	neighbors := []struct {
		region1    game.RegionName
		region2    game.RegionName
		river      bool
		hasCliffs  bool
		dangerZone game.DangerZone
	}{
		{region1: "Lusía", region2: "Lomone"},
		{region1: "Lusía", region2: "Limbol"},
		{region1: "Lusía", region2: "Leil"},
		{region1: "Lomone", region2: "Limbol"},
		{region1: "Limbol", region2: "Leil"},
		{region1: "Limbol", region2: "Worp"},
		{region1: "Leil", region2: "Worp"},
		{region1: "Leil", region2: "Winde"},
		{region1: "Leil", region2: "Ovo", river: true},
		{region1: "Worp", region2: "Winde"},
		{region1: "Worp", region2: "Mare Gond"},
		{region1: "Winde", region2: "Mare Gond"},
		{region1: "Winde", region2: "Mare Elle"},
		{region1: "Winde", region2: "Ovo", river: true},
		{region1: "Ovo", region2: "Mare Elle"},
		{region1: "Zona", region2: "Mare Elle"},
		{region1: "Zona", region2: "Mare Gond"},
		{region1: "Tond", region2: "Tige", dangerZone: "Bankene"},
		{region1: "Tond", region2: "Mare Elle"},
		{region1: "Tond", region2: "Mare Gond"},
		{region1: "Tond", region2: "Mare Ovond"},
		{region1: "Tige", region2: "Mare Elle"},
		{region1: "Tige", region2: "Mare Ovond"},
		{region1: "Tige", region2: "Tusser"},
		{region1: "Tusser", region2: "Gron", dangerZone: "Shangrila"},
		{region1: "Furie", region2: "Firril"},
		{region1: "Furie", region2: "Mare Ovond"},
		{region1: "Firril", region2: "Fond"},
		{region1: "Firril", region2: "Gron"},
		{region1: "Firril", region2: "Gnade"},
		{region1: "Firril", region2: "Mare Ovond"},
		{region1: "Fond", region2: "Mare Ovond"},
		{region1: "Fond", region2: "Mare Unna"},
		{region1: "Gron", region2: "Gnade"},
		{region1: "Gron", region2: "Gewel"},
		{region1: "Gron", region2: "Emman"},
		{region1: "Gnade", region2: "Gewel"},
		{region1: "Gewel", region2: "Mare Unna"},
		{region1: "Gewel", region2: "Emman", hasCliffs: true},
		{region1: "Emman", region2: "Erren", hasCliffs: true},
		{region1: "Emman", region2: "Mare Unna"},
		{region1: "Erren", region2: "Mare Bøso"},
		{region1: "Mare Gond", region2: "Mare Elle"},
		{region1: "Mare Gond", region2: "Mare Ovond"},
		{region1: "Mare Elle", region2: "Mare Ovond", dangerZone: "Bankene"},
		{region1: "Mare Ovond", region2: "Mare Unna"},
		{region1: "Mare Unna", region2: "Mare Bøso"},
	}

	for _, region := range regions {
		board[region.Name] = region
	}

	for _, neighbor := range neighbors {
		region1 := board[neighbor.region1]
		region2 := board[neighbor.region2]

		region1.Neighbors = append(region1.Neighbors, game.Neighbor{
			Name:          neighbor.region2,
			IsAcrossWater: neighbor.river || (region1.IsSea && !region2.IsSea),
			HasCliffs:     neighbor.hasCliffs,
			DangerZone:    neighbor.dangerZone,
		})
		board[neighbor.region1] = region1

		region2.Neighbors = append(region2.Neighbors, game.Neighbor{
			Name:          neighbor.region1,
			IsAcrossWater: neighbor.river || (region2.IsSea && !region1.IsSea),
			HasCliffs:     neighbor.hasCliffs,
			DangerZone:    neighbor.dangerZone,
		})
		board[neighbor.region2] = region2
	}

	return game.New(board, "Test game", 5, MockMessenger{}, log.Default())
}

func placeUnits(units map[game.RegionName]game.Unit, board game.Board) {
	for regionName, unit := range units {
		region := board[regionName]
		region.Unit = unit
		region.ControllingFaction = unit.Faction
		board[regionName] = region
	}
}

func placeOrders(orders []game.Order, board game.Board) {
	for i, order := range orders {
		region, ok := board[order.Origin]
		if !ok {
			continue
		}

		order.Unit = region.Unit
		order.Faction = region.Unit.Faction
		orders[i] = order
	}
}

type ExpectedControl map[game.RegionName]struct {
	ControllingFaction game.PlayerFaction
	Unit               game.Unit
}

func (expected ExpectedControl) check(board game.Board, t *testing.T) {
	for name, region := range board {
		expectation, ok := expected[name]
		if !ok {
			continue
		}

		if region.ControllingFaction != expectation.ControllingFaction {
			t.Errorf(
				"unexpected control of %v, want %v, got %v",
				name,
				region.ControllingFaction,
				expectation.ControllingFaction,
			)
		}
		if region.Unit != expectation.Unit {
			t.Errorf("unexpected unit in %v, want %v, got %v", name, region.Unit, expectation.Unit)
		}
	}
}

type MockMessenger struct{}

func (MockMessenger) SendError(to game.PlayerFaction, err error) {
}

func (MockMessenger) SendOrderRequest(to game.PlayerFaction) error {
	return nil
}

func (MockMessenger) AwaitOrders(
	from game.PlayerFaction,
) ([]game.Order, error) {
	return nil, nil
}

func (MockMessenger) SendOrdersReceived(orders map[game.PlayerFaction][]game.Order) error {
	return nil
}

func (MockMessenger) SendOrdersConfirmation(factionThatSubmittedOrders game.PlayerFaction) error {
	return nil

}
func (MockMessenger) SendSupportRequest(
	to game.PlayerFaction,
	supporting game.RegionName,
	embattled game.RegionName,
	supportable []game.PlayerFaction,
) error {
	return nil
}

func (MockMessenger) AwaitSupport(
	from game.PlayerFaction,
	supporting game.RegionName,
	embattled game.RegionName,
) (supported game.PlayerFaction, err error) {
	return "", nil
}

func (MockMessenger) SendBattleResults(battles []game.Battle) error {
	return nil
}

func (MockMessenger) SendWinner(winner game.PlayerFaction) error {
	return nil
}
