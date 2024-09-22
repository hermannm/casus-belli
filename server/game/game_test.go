package game

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
	"hermannm.dev/opt"
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
				"Erren": movedFrom{"Emman"},
				"Emman": empty,
			},
		},
		{
			name: "SingleplayerBattle",
			units: unitMap{
				"Furie": {Type: UnitKnight, Faction: black},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Furie", Destination: "Firril"},
			},
			expected: expectedUnits{
				"Furie":  stayed,
				"Firril": empty,
			},
		},
		{
			name: "SingleplayerBattleWithSupport",
			units: unitMap{
				"Furie":      {Type: UnitKnight, Faction: black},
				"Mare Ovond": {Type: UnitShip, Faction: black},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Furie", Destination: "Firril"},
				{Type: OrderSupport, Origin: "Mare Ovond", Destination: "Firril"},
			},
			expected: expectedUnits{
				"Furie":  empty,
				"Firril": movedFrom{"Furie"},
			},
		},
		{
			name: "MultiplayerBattleNoDefender",
			units: unitMap{
				"Gron":  {Type: UnitFootman, Faction: white},
				"Gewel": {Type: UnitKnight, Faction: black},
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
				"Gnade": movedFrom{"Gron"},
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
				"Lusía":  stayed,
				"Lomone": empty,
			},
		},
		{
			name: "BorderBattle",
			units: unitMap{
				"Tusser": {Type: UnitFootman, Faction: white},
				"Tige":   {Type: UnitKnight, Faction: black},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Tusser", Destination: "Tige"},
				{Type: OrderMove, Origin: "Tige", Destination: "Tusser"},
			},
			expected: expectedUnits{
				"Tige":   movedFrom{"Tusser"},
				"Tusser": empty,
			},
		},
		{
			name: "UncontestedKnightMove",
			units: unitMap{
				"Lomone": {Type: UnitKnight, Faction: red},
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
				"Worp":   movedFrom{"Lomone"},
				"Limbol": empty,
				"Lomone": empty,
			},
		},
		{
			name: "KnightMovesCuttingSupport",
			units: unitMap{
				"Worp":      {Type: UnitKnight, Faction: green},
				"Winde":     {Type: UnitKnight, Faction: green},
				"Lomone":    {Type: UnitKnight, Faction: red},
				"Lusía":     {Type: UnitKnight, Faction: red},
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
				"Lomone": movedFrom{"Worp"},
				"Lusía":  movedFrom{"Winde"},
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
				"Zona":      movedFrom{"Ovo"},
				"Ovo":       empty,
				"Mare Elle": stayed,
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
				"Fond":       movedFrom{"Winde"},
				"Winde":      empty,
				"Mare Ovond": stayed,
				"Mare Gond":  stayed,
				"Mare Unna":  stayed,
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
				"Leil":   movedFrom{"Worp"},
				"Limbol": movedFrom{"Leil"},
				"Worp":   movedFrom{"Limbol"},
			},
		},
		{
			name: "UncontestedUnitSwap",
			units: unitMap{
				"Dordel": {Type: UnitFootman, Faction: white},
				"Dalom":  {Type: UnitShip, Faction: white},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Dordel", Destination: "Dalom"},
				{Type: OrderMove, Origin: "Dalom", Destination: "Dordel"},
			},
			expected: expectedUnits{
				"Dordel": movedFrom{"Dalom"},
				"Dalom":  movedFrom{"Dordel"},
			},
		},
		{
			name: "ContestedUnitSwap",
			units: unitMap{
				"Firril": {Type: UnitShip, Faction: green},
				"Fond":   {Type: UnitKnight, Faction: green},
				"Gron":   {Type: UnitFootman, Faction: black},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Firril", Destination: "Fond"},
				{Type: OrderMove, Origin: "Fond", Destination: "Firril"},
				{Type: OrderMove, Origin: "Gron", Destination: "Firril"},
			},
			expected: expectedUnits{
				"Fond":   movedFrom{"Firril"},
				"Firril": movedFrom{"Gron"},
				"Gron":   empty,
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

func TestWinterOrders(t *testing.T) {
	testCases := []struct {
		name     string
		units    unitMap
		control  controlMap
		orders   []Order
		expected expectedUnits
	}{
		{
			name: "Build",
			units: unitMap{
				"Calis": {Type: UnitFootman, Faction: yellow},
			},
			control: controlMap{
				"Cymere": yellow,
			},
			orders: []Order{
				{Type: OrderBuild, Origin: "Cymere", UnitType: UnitShip},
				{Type: OrderBuild, Origin: "Pesth", UnitType: UnitKnight},
			},
			expected: expectedUnits{
				"Cymere": Unit{Type: UnitShip, Faction: yellow},
				"Pesth":  Unit{Type: UnitKnight, Faction: yellow},
			},
		},
		{
			name: "Disband",
			units: unitMap{
				"Monté":  {Type: UnitFootman, Faction: red},
				"Brodo":  {Type: UnitFootman, Faction: red},
				"Bassas": {Type: UnitFootman, Faction: red},
				"Bom":    {Type: UnitKnight, Faction: yellow},
			},
			orders: []Order{
				{Type: OrderDisband, Origin: "Brodo"},
			},
			expected: expectedUnits{
				"Brodo":  empty,
				"Monté":  stayed,
				"Bassas": stayed,
				"Bom":    stayed,
			},
		},
		{
			name: "WinterMove",
			units: unitMap{
				"Worp":  {Type: UnitFootman, Faction: green},
				"Winde": {Type: UnitShip, Faction: green},
			},
			control: controlMap{
				"Limbol": green,
				"Lomone": green,
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Worp", Destination: "Lomone"},
				{Type: OrderMove, Origin: "Winde", Destination: "Worp"},
			},
			expected: expectedUnits{
				"Lomone": movedFrom{"Worp"},
				"Worp":   movedFrom{"Winde"},
				"Winde":  empty,
			},
		},
		{
			name: "WinterMoveAfterDisband",
			units: unitMap{
				"Erren":  {Type: UnitShip, Faction: black},
				"Samoje": {Type: UnitFootman, Faction: black},
				"Emman":  {Type: UnitKnight, Faction: white},
			},
			orders: []Order{
				{Type: OrderDisband, Origin: "Erren"},
				{Type: OrderMove, Origin: "Samoje", Destination: "Erren"},
			},
			expected: expectedUnits{
				"Erren":  movedFrom{"Samoje"},
				"Samoje": empty,
			},
		},
		{
			name: "WinterMoveCycle",
			units: unitMap{
				"Calis":  {Type: UnitShip, Faction: yellow},
				"Cymere": {Type: UnitKnight, Faction: yellow},
				"Pesth":  {Type: UnitFootman, Faction: yellow},
			},
			orders: []Order{
				{Type: OrderMove, Origin: "Calis", Destination: "Cymere"},
				{Type: OrderMove, Origin: "Cymere", Destination: "Pesth"},
				{Type: OrderMove, Origin: "Pesth", Destination: "Calis"},
			},
			expected: expectedUnits{
				"Calis":  movedFrom{"Pesth"},
				"Cymere": movedFrom{"Calis"},
				"Pesth":  movedFrom{"Cymere"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			game, board := newMockGame(t, test.units, test.control, test.orders, SeasonWinter)
			game.resolveWinterOrders(test.orders)
			test.expected.check(t, board, test.units)
		})
	}
}

func BenchmarkBoardResolve(b *testing.B) {
	for range b.N {
		b.StopTimer()
		game, orders := benchmarkSetup(b)
		b.StartTimer()
		game.resolveNonWinterOrders(orders)
	}
}

func benchmarkSetup(b *testing.B) (*Game, []Order) {
	units := unitMap{
		"Emman": {Type: UnitFootman, Faction: white},

		"Furie": {Type: UnitKnight, Faction: black},

		"Gron":  {Type: UnitFootman, Faction: white},
		"Gewel": {Type: UnitKnight, Faction: black},

		"Lomone": {Type: UnitFootman, Faction: green},
		"Lusía":  {Type: UnitFootman, Faction: red},
		"Brodo":  {Type: UnitFootman, Faction: red},

		"Tusser": {Type: UnitFootman, Faction: white},
		"Tige":   {Type: UnitKnight, Faction: black},

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
	logHandler := devlog.NewHandler(
		os.Stdout,
		&devlog.Options{Level: slog.LevelDebug, ForceColors: true},
	)
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
	ordersByFaction := make(map[PlayerFaction][]Order)

	for regionName, unit := range units {
		region, ok := board[regionName]
		if !ok {
			t.Fatalf("unit map contained region '%s' not found on board", regionName)
		}

		region.Unit.Put(unit)
		if !region.Sea {
			region.ControllingFaction = unit.Faction
		}

		ordersByFaction[unit.Faction] = nil
	}

	for regionName, faction := range control {
		region, ok := board[regionName]
		if !ok {
			t.Fatalf("control map contained region '%s' not found on board", regionName)
		}

		region.ControllingFaction = faction

		ordersByFaction[faction] = nil
	}

	for i, order := range orders {
		region, ok := board[order.Origin]
		if !ok {
			t.Fatalf("order origin region '%s' not found on board", order.Origin)
		}

		if order.Type == OrderBuild {
			order.Faction = region.ControllingFaction
		} else if region.Unit.HasValue() {
			order.Faction = region.Unit.Value.Faction
			order.UnitType = region.Unit.Value.Type
		}
		orders[i] = order

		ordersByFaction[order.Faction] = append(ordersByFaction[order.Faction], order)
	}

	for faction, orders := range ordersByFaction {
		if err := validateOrders(orders, faction, board, season); err != nil {
			t.Fatal(wrap.Error(err, "invalid orders in test setup"))
		}
	}

	return New(board, baseBoardInfo, MockMessenger{}, log.Default(), diceRollerForTests), board
}

// Maps region names to either a Unit (which may be empty), a movedFrom struct, or stayed.
type expectedUnits map[RegionName]any

var (
	empty  Unit
	stayed struct{}
)

// Region name from the original unit map where we expect a unit to have moved from.
type movedFrom struct {
	region RegionName
}

func (expected expectedUnits) check(t *testing.T, board Board, originalUnits unitMap) {
	for regionName, expected := range expected {
		region, ok := board[regionName]
		if !ok {
			t.Fatalf("invalid test setup: '%s' is not a region on the board", regionName)
		}

		var expectedUnit opt.Option[Unit]
		switch expected := expected.(type) {
		case Unit:
			if expected != empty {
				expectedUnit.Put(expected)
			}
		case movedFrom:
			unit, ok := originalUnits[expected.region]
			if !ok {
				t.Fatalf("invalid test setup: no unit for region '%s' in unit map", expected)
			}
			expectedUnit.Put(unit)
		case struct{}: // stayed
			unit, ok := originalUnits[regionName]
			if !ok {
				t.Fatalf(
					"invalid test setup: no staying unit for region '%s' in unit map",
					regionName,
				)
			}
			expectedUnit.Put(unit)
		default:
			t.Fatalf(
				"invalid test setup: expectedUnit in region '%s' is not Unit, movedFrom or stayed",
				regionName,
			)
		}

		if region.Unit != expectedUnit {
			t.Errorf("%s: want %v, got %v", regionName, expectedUnit, region.Unit)
		}
	}
}

type MockMessenger struct{}

func (MockMessenger) SendError(to PlayerFaction, err error) {}
func (MockMessenger) SendGameStarted(board Board)           {}
func (MockMessenger) SendOrderRequest(to PlayerFaction, season Season) (succeeded bool) {
	return true
}
func (MockMessenger) SendOrdersReceived(orders map[PlayerFaction][]Order)             {}
func (MockMessenger) SendOrdersConfirmation(factionThatSubmittedOrders PlayerFaction) {}
func (MockMessenger) SendBattleAnnouncement(battle Battle)                            {}
func (MockMessenger) SendBattleResults(battle Battle)                                 {}
func (MockMessenger) SendWinner(winner PlayerFaction)                                 {}
func (MockMessenger) AwaitOrders(ctx context.Context, from PlayerFaction) ([]Order, error) {
	return nil, nil
}
func (MockMessenger) AwaitSupport(
	ctx context.Context,
	from PlayerFaction,
	embattled RegionName,
) (supported PlayerFaction, err error) {
	return "", nil
}
func (MockMessenger) AwaitDiceRoll(ctx context.Context, from PlayerFaction) error {
	return nil
}
func (MockMessenger) ClearMessages() {}
