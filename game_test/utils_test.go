package game_test

import (
	"log/slog"
	"os"
	"testing"

	"hermannm.dev/bfh-server/game"
	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
)

func TestMain(m *testing.M) {
	os.Setenv("FORCE_COLOR", "1")
	log.ColorsEnabled = true
	logHandler := devlog.NewHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(logHandler))

	os.Exit(m.Run())
}

func diceRollerForTests() int {
	return 3
}

const (
	yellow game.PlayerFaction = "Yellow"
	red    game.PlayerFaction = "Red"
	green  game.PlayerFaction = "Green"
	white  game.PlayerFaction = "White"
	black  game.PlayerFaction = "Black"
)

type unitMap map[game.RegionName]game.Unit
type controlMap map[game.RegionName]game.PlayerFaction

func newMockGame(
	t testing.TB,
	units unitMap,
	control controlMap,
	orders []game.Order,
	season game.Season,
) *game.Game {
	regions := []*game.Region{
		{Name: "Bom"},
		{Name: "Brodo", Forest: true},
		{Name: "Bassas", Forest: true, Castle: true},
		{Name: "Lusía", Castle: true},
		{Name: "Lomone", Forest: true},
		{Name: "Limbol", Forest: true},
		{Name: "Leil"},
		{Name: "Worp", Forest: true, HomeFaction: green, ControllingFaction: green},
		{
			Name:               "Winde",
			Forest:             true,
			Castle:             true,
			HomeFaction:        green,
			ControllingFaction: green,
		},
		{Name: "Ovo", Forest: true},
		{Name: "Mare Duna", Sea: true},
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
		{Name: "Emman", Forest: true, HomeFaction: black, ControllingFaction: black},
		{Name: "Erren", Castle: true, HomeFaction: black, ControllingFaction: black},
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
		{region1: "Lomone", region2: "Mare Duna"},
		{region1: "Limbol", region2: "Leil"},
		{region1: "Limbol", region2: "Worp"},
		{region1: "Limbol", region2: "Mare Duna"},
		{region1: "Leil", region2: "Worp"},
		{region1: "Leil", region2: "Winde"},
		{region1: "Leil", region2: "Ovo", river: true},
		{region1: "Worp", region2: "Winde"},
		{region1: "Worp", region2: "Mare Duna"},
		{region1: "Worp", region2: "Mare Gond"},
		{region1: "Winde", region2: "Mare Gond"},
		{region1: "Winde", region2: "Mare Elle"},
		{region1: "Winde", region2: "Ovo", river: true},
		{region1: "Ovo", region2: "Mare Elle"},
		{region1: "Zona", region2: "Mare Elle"},
		{region1: "Zona", region2: "Mare Gond"},
		{region1: "Tond", region2: "Tige", dangerZone: "Bankene", river: true},
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
		{region1: "Mare Duna", region2: "Mare Gond"},
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

	for regionName, unit := range units {
		region := board[regionName]
		region.Unit = unit
		if !region.Sea {
			region.ControllingFaction = unit.Faction
		}
	}

	for regionName, faction := range control {
		region := board[regionName]
		region.ControllingFaction = faction
	}

	ordersByFaction := make(map[game.PlayerFaction][]game.Order)
	for i, order := range orders {
		region, ok := board[order.Origin]
		if !ok {
			continue
		}

		order.UnitType = region.Unit.Type
		order.Faction = region.Unit.Faction
		orders[i] = order

		ordersByFaction[order.Faction] = append(ordersByFaction[order.Faction], order)
	}

	for _, orders := range ordersByFaction {
		if err := game.ValidateOrders(orders, board, season); err != nil {
			t.Fatal(wrap.Error(err, "invalid orders in test setup"))
		}
	}

	boardInfo := game.BoardInfo{ID: "test", Name: "Test game", WinningCastleCount: 5}
	return game.New(board, boardInfo, MockMessenger{}, log.Default(), diceRollerForTests)
}

type expectedUnits map[game.RegionName]game.RegionName

func (expected expectedUnits) check(t *testing.T, testGame *game.Game, units unitMap) {
	for regionName, expected := range expected {
		region, ok := testGame.GetBoardRegion(regionName)
		if !ok {
			t.Fatalf("invalid test setup: '%s' is not a region on the board", regionName)
		}

		var expectedUnit game.Unit
		if expected != "" {
			unit, ok := units[expected]
			if !ok {
				t.Fatalf("invalid test setup: no unit for region '%s' in unit map", expected)
			}
			expectedUnit = unit
		}

		if region.Unit != expectedUnit {
			if expected == "" {
				t.Errorf("%s: want no unit, got %v", regionName, region.Unit)
			} else if region.Unit.IsNone() {
				t.Errorf("%s: want %v, got no unit", regionName, expectedUnit)
			} else {
				t.Errorf(
					"%s: want %v, got %v",
					regionName,
					expectedUnit,
					region.Unit,
				)
			}
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

func (MockMessenger) AwaitOrders(from game.PlayerFaction) ([]game.Order, error) {
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

func (MockMessenger) SendDangerZoneCrossings(crossings []game.DangerZoneCrossing) error {
	return nil
}

func (MockMessenger) SendWinner(winner game.PlayerFaction) error {
	return nil
}

func (MockMessenger) ClearMessages() {}
