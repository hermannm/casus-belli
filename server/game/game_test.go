package game

import (
	"log/slog"
	"os"
	"testing"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
)

func TestNonWinterOrders(t *testing.T) {
	testCases := []struct {
		name     string
		units    unitMap
		control  controlMap
		orders   []Order
		expected expectedUnits
	}{
		{
			name: "UncontestedMove",
			units: unitMap{
				"Emman": {Type: UnitFootman, Faction: white},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Emman", Destination: "Erren"},
			},
			expected: expectedUnits{
				"Erren": "Emman",
				"Emman": empty,
			},
		},
		{
			name: "SingleplayerBattle",
			units: unitMap{
				"Furie": {Type: UnitHorse, Faction: black},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Furie", Destination: "Firril"},
			},
			expected: expectedUnits{
				"Furie":  "Furie",
				"Firril": empty,
			},
		},
		{
			name: "SingleplayerBattleWithSupport",
			units: unitMap{
				"Furie":      {Type: UnitHorse, Faction: black},
				"Mare Ovond": {Type: UnitShip, Faction: black},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Furie", Destination: "Firril"},
				{Type: OrderSupport, Origin: "Mare Ovond", Destination: "Firril"},
			},
			expected: expectedUnits{
				"Furie":  empty,
				"Firril": "Furie",
			},
		},
		{
			name: "MultiplayerBattleNoDefender",
			units: unitMap{
				"Gron":  {Type: UnitFootman, Faction: white},
				"Gewel": {Type: UnitHorse, Faction: black},
			},
			control: controlMap{
				"Gnade": black,
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Gron", Destination: "Gnade"},
				{Type: OrderMove, Origin: "Gewel", Destination: "Gnade"},
			},
			expected: expectedUnits{
				"Gron":  empty,
				"Gewel": empty,
				"Gnade": "Gron",
			},
		},
		{
			name: "MultiplayerBattleWithSupportedDefender",
			units: unitMap{
				"Lomone": {Type: UnitFootman, Faction: green},
				"Lusía":  {Type: UnitFootman, Faction: red},
				"Brodo":  {Type: UnitFootman, Faction: red},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Lomone", Destination: "Lusía"},
				{Type: OrderSupport, Origin: "Brodo", Destination: "Lusía"},
			},
			expected: expectedUnits{
				"Lusía":  "Lusía",
				"Lomone": empty,
			},
		},
		{
			name: "BorderBattle",
			units: unitMap{
				"Tusser": {Type: UnitFootman, Faction: white},
				"Tige":   {Type: UnitHorse, Faction: black},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Tusser", Destination: "Tige"},
				{Type: OrderMove, Origin: "Tige", Destination: "Tusser"},
			},
			expected: expectedUnits{
				"Tige":   "Tusser",
				"Tusser": empty,
			},
		},
		{
			name: "UncontestedHorseMove",
			units: unitMap{
				"Lomone": {Type: UnitHorse, Faction: red},
			},
			control: controlMap{
				"Limbol": red,
				"Worp":   green,
			},
			orders: []Order{
				{
					Type:              OrderMove,
					Origin:            "Lomone",
					Destination:       "Limbol",
					SecondDestination: "Worp",
				},
			},
			expected: expectedUnits{
				"Worp":   "Lomone",
				"Limbol": empty,
				"Lomone": empty,
			},
		},
		{
			name: "HorseMovesCuttingSupport",
			units: unitMap{
				"Worp":      {Type: UnitHorse, Faction: green},
				"Winde":     {Type: UnitHorse, Faction: green},
				"Lomone":    {Type: UnitHorse, Faction: red},
				"Lusía":     {Type: UnitHorse, Faction: red},
				"Mare Illa": {Type: UnitShip, Faction: green},
				"Mare Duna": {Type: UnitShip, Faction: green},
				"Morone":    {Type: UnitFootman, Faction: green},
				"Brodo":     {Type: UnitFootman, Faction: green},
			},
			control: controlMap{
				"Limbol": red,
				"Leil":   red,
			},
			orders: []Order{
				{
					Type:              OrderMove,
					Origin:            "Worp",
					Destination:       "Limbol",
					SecondDestination: "Lomone",
				},
				{
					Type:              OrderMove,
					Origin:            "Winde",
					Destination:       "Leil",
					SecondDestination: "Lusía",
				},
				{Type: OrderSupport, Origin: "Lusía", Destination: "Lomone"},
				{Type: OrderSupport, Origin: "Lomone", Destination: "Lusía"},
				{Type: OrderSupport, Origin: "Mare Illa", Destination: "Lomone"},
				{Type: OrderSupport, Origin: "Mare Duna", Destination: "Lomone"},
				{Type: OrderSupport, Origin: "Morone", Destination: "Lusía"},
				{Type: OrderSupport, Origin: "Brodo", Destination: "Lusía"},
			},
			expected: expectedUnits{
				"Lomone": "Worp",
				"Lusía":  "Winde",
				"Limbol": empty,
				"Leil":   empty,
				"Worp":   empty,
				"Winde":  empty,
			},
		},
		{
			name: "Transport",
			units: unitMap{
				"Ovo":       {Type: UnitFootman, Faction: green},
				"Mare Elle": {Type: UnitShip, Faction: green},
			},
			control: controlMap{
				"Zona": white,
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Ovo", Destination: "Zona"},
				{Type: OrderTransport, Origin: "Mare Elle"},
			},
			expected: expectedUnits{
				"Zona":      "Ovo",
				"Ovo":       empty,
				"Mare Elle": "Mare Elle",
			},
		},
		{
			name: "TransportAttacked",
			units: unitMap{
				"Winde":      {Type: UnitFootman, Faction: green},
				"Mare Gond":  {Type: UnitShip, Faction: green},
				"Mare Ovond": {Type: UnitShip, Faction: green},
				"Mare Unna":  {Type: UnitShip, Faction: black},
			},
			control: controlMap{
				"Fond": black,
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Winde", Destination: "Fond"},
				{Type: OrderTransport, Origin: "Mare Gond"},
				{Type: OrderTransport, Origin: "Mare Ovond"},
				{Type: OrderMove, Origin: "Mare Unna", Destination: "Mare Ovond"},
			},
			expected: expectedUnits{
				"Fond":       "Winde",
				"Winde":      empty,
				"Mare Ovond": "Mare Ovond",
				"Mare Gond":  "Mare Gond",
				"Mare Unna":  "Mare Unna",
			},
		},
		{
			name: "UncontestedMoveCycle",
			units: unitMap{
				"Leil":   {Type: UnitFootman, Faction: red},
				"Limbol": {Type: UnitFootman, Faction: green},
				"Worp":   {Type: UnitFootman, Faction: yellow},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Leil", Destination: "Limbol"},
				{Type: OrderMove, Origin: "Limbol", Destination: "Worp"},
				{Type: OrderMove, Origin: "Worp", Destination: "Leil"},
			},
			expected: expectedUnits{
				"Leil":   "Worp",
				"Limbol": "Leil",
				"Worp":   "Limbol",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			game, board := newMockGame(t, test.units, test.control, test.orders, SeasonSpring)
			game.resolveNonWinterOrders(test.orders)
			test.expected.check(t, board, test.units)
		})
	}
}

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		game, orders := benchmarkSetup(b)
		b.StartTimer()
		game.resolveNonWinterOrders(orders)
	}
}

func benchmarkSetup(b *testing.B) (*Game, []Order) {
	units := unitMap{
		"Emman": {Type: UnitFootman, Faction: white},

		"Furie": {Type: UnitHorse, Faction: black},

		"Gron":  {Type: UnitFootman, Faction: white},
		"Gewel": {Type: UnitHorse, Faction: black},

		"Lomone": {Type: UnitFootman, Faction: green},
		"Lusía":  {Type: UnitFootman, Faction: red},
		"Brodo":  {Type: UnitFootman, Faction: red},

		"Tusser": {Type: UnitFootman, Faction: white},
		"Tige":   {Type: UnitHorse, Faction: black},

		"Tond": {Type: UnitFootman, Faction: green},

		"Ovo":       {Type: UnitFootman, Faction: green},
		"Mare Elle": {Type: UnitShip, Faction: green},

		"Winde":      {Type: UnitFootman, Faction: green},
		"Mare Gond":  {Type: UnitShip, Faction: green},
		"Mare Ovond": {Type: UnitShip, Faction: green},
		"Mare Unna":  {Type: UnitShip, Faction: black},

		"Leil":   {Type: UnitFootman, Faction: red},
		"Limbol": {Type: UnitFootman, Faction: green},
		"Worp":   {Type: UnitFootman, Faction: yellow},
	}

	orders := []Order{
		// Auto-success
		{Type: OrderMove, Origin: "Emman", Destination: "Erren"},

		// Singleplayer battle
		{Type: OrderMove, Origin: "Furie", Destination: "Firril"},

		// Multiplayer battle, no defender
		{Type: OrderMove, Origin: "Gron", Destination: "Gnade"},
		{Type: OrderMove, Origin: "Gewel", Destination: "Gnade"},

		// Multiplayer battle with supported defender
		{Type: OrderMove, Origin: "Lomone", Destination: "Lusía"},
		{Type: OrderSupport, Origin: "Brodo", Destination: "Lusía"},

		// Border battle
		{Type: OrderMove, Origin: "Tusser", Destination: "Tige"},
		{Type: OrderMove, Origin: "Tige", Destination: "Tusser"},

		// Danger zone, dependent move
		{Type: OrderMove, Origin: "Tond", Destination: "Tige"},

		// Transport
		{Type: OrderMove, Origin: "Ovo", Destination: "Zona"},
		{Type: OrderTransport, Origin: "Mare Elle"},

		// Transport attacked
		{Type: OrderMove, Origin: "Winde", Destination: "Fond"},
		{Type: OrderTransport, Origin: "Mare Gond"},
		{Type: OrderTransport, Origin: "Mare Ovond"},
		{Type: OrderMove, Origin: "Mare Unna", Destination: "Mare Ovond"},

		// Move cycle
		{Type: OrderMove, Origin: "Leil", Destination: "Limbol"},
		{Type: OrderMove, Origin: "Limbol", Destination: "Worp"},
		{Type: OrderMove, Origin: "Worp", Destination: "Leil"},
	}

	game, _ := newMockGame(b, units, nil, orders, SeasonSpring)
	return game, orders
}

func TestMain(m *testing.M) {
	os.Setenv("FORCE_COLOR", "1")
	log.ColorsEnabled = true
	logHandler := devlog.NewHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(logHandler))

	board, boardInfo, err := ReadBoardFromConfigFile("casus-belli-5players")
	if err != nil {
		log.ErrorCause(err, "failed to read board config for tests")
		os.Exit(1)
	}
	emptyBoard, baseBoardInfo = board, boardInfo

	os.Exit(m.Run())
}

var (
	emptyBoard    Board
	baseBoardInfo BoardInfo
)

func diceRollerForTests() int {
	return 3
}

const (
	yellow PlayerFaction = "Yellow"
	red    PlayerFaction = "Red"
	green  PlayerFaction = "Green"
	white  PlayerFaction = "White"
	black  PlayerFaction = "Black"

	empty = ""
)

type unitMap map[RegionName]Unit
type controlMap map[RegionName]PlayerFaction

func newMockGame(
	t testing.TB,
	units unitMap,
	control controlMap,
	orders []Order,
	season Season,
) (*Game, Board) {
	board := emptyBoard.copy()

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

	ordersByFaction := make(map[PlayerFaction][]Order)
	for i, order := range orders {
		region, ok := board[order.Origin]
		if !ok {
			t.Fatalf("order origin region '%s' not found on board", order.Origin)
		}

		order.UnitType = region.Unit.Type
		order.Faction = region.Unit.Faction
		orders[i] = order

		ordersByFaction[order.Faction] = append(ordersByFaction[order.Faction], order)
	}

	for _, orders := range ordersByFaction {
		if err := validateOrders(orders, board, season); err != nil {
			t.Fatal(wrap.Error(err, "invalid orders in test setup"))
		}
	}

	return New(board, baseBoardInfo, MockMessenger{}, log.Default(), diceRollerForTests), board
}

// Maps region names to either a Unit, or a region name from the original unit map where we expect
// the unit to have moved from.
type expectedUnits map[RegionName]any

func (expected expectedUnits) check(t *testing.T, board Board, originalUnits unitMap) {
	for regionName, expected := range expected {
		region, ok := board[regionName]
		if !ok {
			t.Fatalf("invalid test setup: '%s' is not a region on the board", regionName)
		}

		var expectedUnit Unit
		switch expected := expected.(type) {
		case Unit:
			expectedUnit = expected
		case string:
			if expected != empty {
				unit, ok := originalUnits[RegionName(expected)]
				if !ok {
					t.Fatalf("invalid test setup: no unit for region '%s' in unit map", expected)
				}
				expectedUnit = unit
			}
		default:
			t.Fatalf("invalid test setup: expectedUnit is not Unit or RegionName")
		}

		if region.Unit != expectedUnit {
			if expectedUnit.isNone() {
				t.Errorf("%s: want no unit, got %v", regionName, region.Unit)
			} else if region.Unit.isNone() {
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

func (MockMessenger) SendError(to PlayerFaction, err error) {}

func (MockMessenger) SendGameStarted(board Board) error {
	return nil
}

func (MockMessenger) SendOrderRequest(to PlayerFaction, season Season) error {
	return nil
}

func (MockMessenger) AwaitOrders(from PlayerFaction) ([]Order, error) {
	return nil, nil
}

func (MockMessenger) SendOrdersReceived(orders map[PlayerFaction][]Order) error {
	return nil
}

func (MockMessenger) SendOrdersConfirmation(factionThatSubmittedOrders PlayerFaction) error {
	return nil

}
func (MockMessenger) SendSupportRequest(
	to PlayerFaction,
	supporting RegionName,
	embattled RegionName,
	supportable []PlayerFaction,
) error {
	return nil
}

func (MockMessenger) AwaitSupport(
	from PlayerFaction,
	supporting RegionName,
	embattled RegionName,
) (supported PlayerFaction, err error) {
	return "", nil
}

func (MockMessenger) SendBattleResults(battles ...Battle) error {
	return nil
}

func (MockMessenger) SendDangerZoneCrossings(crossings []DangerZoneCrossing) error {
	return nil
}

func (MockMessenger) SendWinner(winner PlayerFaction) error {
	return nil
}

func (MockMessenger) ClearMessages() {}
