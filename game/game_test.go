package game

import (
	"log/slog"
	"os"
	"testing"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
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
				"Emman": "",
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
				"Firril": "",
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
				"Furie":  "",
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
				"Gron":  "",
				"Gewel": "",
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
				"Lomone": "",
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
				"Tusser": "",
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
				"Ovo":       "",
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
				"Winde":      "",
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

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			game := newMockGame(t, testCase.units, testCase.control, testCase.orders, SeasonSpring)
			game.resolveNonWinterOrders(testCase.orders)
			testCase.expected.check(t, game, testCase.units)
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

	game := newMockGame(b, units, nil, orders, SeasonSpring)
	return game, orders
}

func TestMain(m *testing.M) {
	os.Setenv("FORCE_COLOR", "1")
	log.ColorsEnabled = true
	logHandler := devlog.NewHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(logHandler))

	os.Exit(m.Run())
}

type unitMap map[RegionName]Unit
type controlMap map[RegionName]PlayerFaction

func newMockGame(
	t testing.TB,
	units unitMap,
	control controlMap,
	orders []Order,
	season Season,
) *Game {
	regions := []*Region{
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
		region1    RegionName
		region2    RegionName
		river      bool
		hasCliffs  bool
		dangerZone DangerZone
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
		{region1: "Mare Gond", region2: "Mare Elle"},
		{region1: "Mare Gond", region2: "Mare Ovond"},
		{region1: "Mare Elle", region2: "Mare Ovond", dangerZone: "Bankene"},
		{region1: "Mare Ovond", region2: "Mare Unna"},
		{region1: "Mare Unna", region2: "Mare Bøso"},
	}

	board := make(Board)
	for _, region := range regions {
		board[region.Name] = region
	}

	for _, neighbor := range neighbors {
		region1 := board[neighbor.region1]
		region2 := board[neighbor.region2]

		region1.Neighbors = append(region1.Neighbors, Neighbor{
			Name:        neighbor.region2,
			AcrossWater: neighbor.river || (region1.Sea && !region2.Sea),
			Cliffs:      neighbor.hasCliffs,
			DangerZone:  neighbor.dangerZone,
		})

		region2.Neighbors = append(region2.Neighbors, Neighbor{
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

	ordersByFaction := make(map[PlayerFaction][]Order)
	for i, order := range orders {
		region, ok := board[order.Origin]
		if !ok {
			continue
		}

		order.unitType = region.Unit.Type
		order.Faction = region.Unit.Faction
		orders[i] = order

		ordersByFaction[order.Faction] = append(ordersByFaction[order.Faction], order)
	}

	for _, orders := range ordersByFaction {
		if err := validateOrders(orders, board, season); err != nil {
			t.Fatal(wrap.Error(err, "invalid orders in test setup"))
		}
	}

	boardInfo := BoardInfo{ID: "test", Name: "Test game", WinningCastleCount: 5}
	return New(board, boardInfo, MockMessenger{}, log.Default(), diceRollerForTests)
}

type expectedUnits map[RegionName]RegionName

func (expected expectedUnits) check(t *testing.T, game *Game, units unitMap) {
	for regionName, expected := range expected {
		region, ok := game.board[regionName]
		if !ok {
			t.Fatalf("invalid test setup: '%s' is not a region on the board", regionName)
		}

		var expectedUnit Unit
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
