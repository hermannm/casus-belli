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

	board, boardInfo, err := game.ReadBoardFromConfigFile("bfh_5players")
	if err != nil {
		log.ErrorCause(err, "failed to read board config for tests")
		os.Exit(1)
	}
	emptyBoard, baseBoardInfo = board, boardInfo

	os.Exit(m.Run())
}

var (
	emptyBoard    game.Board
	baseBoardInfo game.BoardInfo
)

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
) (*game.Game, game.Board) {
	board := emptyBoard.Copy()

	for regionName, unit := range units {
		region, ok := board[regionName]
		if !ok {
			t.Fatalf("unit map contained region '%s' not found on board", regionName)
		}

		region.Unit = unit
		if !region.Sea {
			region.ControllingFaction = unit.Faction
		}
	}

	for regionName, faction := range control {
		region, ok := board[regionName]
		if !ok {
			t.Fatalf("control map contained region '%s' not found on board", regionName)
		}

		region.ControllingFaction = faction
	}

	ordersByFaction := make(map[game.PlayerFaction][]game.Order)
	for i, order := range orders {
		region, ok := board[order.Origin]
		if !ok {
			t.Fatalf("order contained origin region '%s' not found on board", order.Origin)
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

	return game.New(board, baseBoardInfo, MockMessenger{}, log.Default(), diceRollerForTests), board
}

type expectedUnits map[game.RegionName]game.RegionName

func (expected expectedUnits) check(t *testing.T, board game.Board, originalUnits unitMap) {
	for regionName, expected := range expected {
		region, ok := board[regionName]
		if !ok {
			t.Fatalf("invalid test setup: '%s' is not a region on the board", regionName)
		}

		var expectedUnit game.Unit
		if expected != "" {
			unit, ok := originalUnits[expected]
			if !ok {
				t.Fatalf("invalid test setup: no unit for region '%s' in unit map", expected)
			}
			expectedUnit = unit
		}

		if region.Unit != expectedUnit {
			var emptyUnit game.Unit
			if expectedUnit == emptyUnit {
				t.Errorf("%s: want no unit, got %v", regionName, region.Unit)
			} else if region.Unit == emptyUnit {
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
