package orderresolving_test

import (
	"testing"

	"hermannm.dev/bfh-server/game/gametypes"
)

func newMockBoard() gametypes.Board {
	board := gametypes.Board{
		Regions:            make(map[string]gametypes.Region),
		Name:               "Mock board",
		WinningCastleCount: 5,
	}

	regions := []gametypes.Region{
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
		region1    string
		region2    string
		river      bool
		hasCliffs  bool
		dangerZone string
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
		board.Regions[region.Name] = region
	}

	for _, neighbor := range neighbors {
		region1 := board.Regions[neighbor.region1]
		region2 := board.Regions[neighbor.region2]

		region1.Neighbors = append(region1.Neighbors, gametypes.Neighbor{
			Name:          neighbor.region2,
			IsAcrossWater: neighbor.river || (region1.IsSea && !region2.IsSea),
			HasCliffs:     neighbor.hasCliffs,
			DangerZone:    neighbor.dangerZone,
		})
		board.Regions[neighbor.region1] = region1

		region2.Neighbors = append(region2.Neighbors, gametypes.Neighbor{
			Name:          neighbor.region1,
			IsAcrossWater: neighbor.river || (region2.IsSea && !region1.IsSea),
			HasCliffs:     neighbor.hasCliffs,
			DangerZone:    neighbor.dangerZone,
		})
		board.Regions[neighbor.region2] = region2
	}

	return board
}

func placeUnits(units map[string]gametypes.Unit, board gametypes.Board) {
	for regionName, unit := range units {
		region := board.Regions[regionName]
		region.Unit = unit
		region.ControllingFaction = unit.Faction
		board.Regions[regionName] = region
	}
}

func placeOrders(orders []gametypes.Order, board gametypes.Board) {
	for i, order := range orders {
		region, ok := board.Regions[order.Origin]
		if !ok {
			continue
		}

		order.Unit = region.Unit
		order.Faction = region.Unit.Faction
		orders[i] = order
	}
}

type ExpectedControl map[string]struct {
	ControllingFaction gametypes.PlayerFaction
	Unit               gametypes.Unit
}

func (expected ExpectedControl) check(board gametypes.Board, t *testing.T) {
	for name, region := range board.Regions {
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

func (MockMessenger) SendBattleResults(battles []gametypes.Battle) error {
	return nil
}

func (MockMessenger) SendSupportRequest(
	to gametypes.PlayerFaction,
	supportingRegion string,
	embattledRegion string,
	supportableFactions []gametypes.PlayerFaction,
) error {
	return nil
}

func (MockMessenger) AwaitSupport(
	from gametypes.PlayerFaction,
	supportingRegion string,
	embattledRegion string,
) (supported gametypes.PlayerFaction, err error) {
	return "", nil
}
