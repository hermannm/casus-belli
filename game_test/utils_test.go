package game_test

import (
	"testing"

	"hermannm.dev/bfh-server/game"
	"hermannm.dev/devlog/log"
)

func newMockGame() *game.Game {
	regions := []*game.Region{
		{Name: "Bom"},
		{Name: "Brodo", Forest: true},
		{Name: "Bassas", Forest: true, Castle: true},
		{Name: "Lusía", Castle: true},
		{Name: "Lomone", Forest: true},
		{Name: "Limbol", Forest: true},
		{Name: "Leil"},
		{Name: "Worp", Forest: true, HomeFaction: "Green", ControllingFaction: "Green"},
		{
			Name:               "Winde",
			Forest:             true,
			Castle:             true,
			HomeFaction:        "Green",
			ControllingFaction: "Green",
		},
		{Name: "Ovo", Forest: true},
		{Name: "Mare Gond", Sea: true},
		{Name: "Mare Elle", Sea: true},
		{Name: "Zona"},
		{Name: "Tond"},
		{Name: "Tige"},
		{Name: "Tusser"},
		{Name: "Mare Ovond", Sea: true},
		{Name: "Furie", Castle: true},
		{Name: "Firril"},
		{Name: "Fond"},
		{Name: "Gron"},
		{Name: "Gnade"},
		{Name: "Gewel", Forest: true, Castle: true},
		{Name: "Mare Unna", Sea: true},
		{Name: "Emman", Forest: true, HomeFaction: "Black", ControllingFaction: "Black"},
		{Name: "Erren", Castle: true, HomeFaction: "Black", ControllingFaction: "Black"},
		{Name: "Mare Bøso", Sea: true},
	}

	neighbors := []struct {
		region1    game.RegionName
		region2    game.RegionName
		river      bool
		hasCliffs  bool
		dangerZone game.DangerZone
	}{
		{region1: "Bom", region2: "Brodo"},
		{region1: "Bom", region2: "Bassas"},
		{region1: "Brodo", region2: "Bassas"},
		{region1: "Brodo", region2: "Lusía"},
		{region1: "Brodo", region2: "Leil"},
		{region1: "Bassas", region2: "Leil"},
		{region1: "Bassas", region2: "Ovo"},
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

	board := make(game.Board)
	for _, region := range regions {
		board[region.Name] = region
	}

	for _, neighbor := range neighbors {
		region1 := board[neighbor.region1]
		region2 := board[neighbor.region2]

		region1.Neighbors = append(region1.Neighbors, game.Neighbor{
			Name:        neighbor.region2,
			AcrossWater: neighbor.river || (region1.Sea && !region2.Sea),
			Cliffs:      neighbor.hasCliffs,
			DangerZone:  neighbor.dangerZone,
		})

		region2.Neighbors = append(region2.Neighbors, game.Neighbor{
			Name:        neighbor.region1,
			AcrossWater: neighbor.river || (region2.Sea && !region1.Sea),
			Cliffs:      neighbor.hasCliffs,
			DangerZone:  neighbor.dangerZone,
		})
	}

	boardInfo := game.BoardInfo{ID: "test", Name: "Test game", WinningCastleCount: 5}
	return game.New(board, boardInfo, MockMessenger{}, log.Default())
}

func placeUnits(units map[game.RegionName]game.Unit, board game.Board) {
	for regionName, unit := range units {
		region := board[regionName]
		region.Unit = unit
		region.ControllingFaction = unit.Faction
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

func (MockMessenger) SendError(to game.PlayerFaction, err error) {}

func (MockMessenger) SendGameStarted(board game.Board) error {
	return nil
}

func (MockMessenger) SendOrderRequest(to game.PlayerFaction, season game.Season) error {
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

func (MockMessenger) SendBattleResults(battles ...game.Battle) error {
	return nil
}

func (MockMessenger) SendWinner(winner game.PlayerFaction) error {
	return nil
}

func (MockMessenger) ClearMessages() {}
