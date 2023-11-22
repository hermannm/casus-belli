package game

import (
	"testing"

	"hermannm.dev/devlog/log"
)

// Tests whether units correctly move in circle without outside interference.
func TestResolveConflictFreeMoveCycle(t *testing.T) {
	units := map[RegionName]Unit{
		"Leil":   {Type: UnitFootman, Faction: "Red"},
		"Limbol": {Type: UnitFootman, Faction: "Green"},
		"Worp":   {Type: UnitFootman, Faction: "Yellow"},
	}

	orders := []Order{
		{Type: OrderMove, Origin: "Leil", Destination: "Limbol"},
		{Type: OrderMove, Origin: "Limbol", Destination: "Worp"},
		{Type: OrderMove, Origin: "Worp", Destination: "Leil"},
	}

	game := newMockGame()
	placeUnits(units, game.board)
	placeOrders(orders, game.board)

	game.resolveNonWinterOrders(orders)

	ExpectedControl{
		"Leil":   {ControllingFaction: "Yellow", Unit: units["Worp"]},
		"Limbol": {ControllingFaction: "Red", Unit: units["Leil"]},
		"Worp":   {ControllingFaction: "Green", Unit: units["Limbol"]},
	}.check(game.board, t)
}

func BenchmarkBoardResolve(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		game, orders := benchmarkSetup()
		b.StartTimer()
		game.resolveNonWinterOrders(orders)
	}
}

func newMockGame() *Game {
	regions := []*Region{
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

	boardInfo := BoardInfo{ID: "test", Name: "Test game", WinningCastleCount: 5}
	return New(board, boardInfo, MockMessenger{}, log.Default())
}

func placeUnits(units map[RegionName]Unit, board Board) {
	for regionName, unit := range units {
		region := board[regionName]
		region.Unit = unit
		region.ControllingFaction = unit.Faction
	}
}

func placeOrders(orders []Order, board Board) {
	for i, order := range orders {
		region, ok := board[order.Origin]
		if !ok {
			continue
		}

		order.unit = region.Unit
		order.Faction = region.Unit.Faction
		orders[i] = order
	}
}

type ExpectedControl map[RegionName]struct {
	ControllingFaction PlayerFaction
	Unit               Unit
}

func (expected ExpectedControl) check(board Board, t *testing.T) {
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

func benchmarkSetup() (*Game, []Order) {
	units := map[RegionName]Unit{
		"Emman": {Type: UnitFootman, Faction: "White"},

		"Lomone": {Type: UnitFootman, Faction: "Green"},
		"Lusía":  {Type: UnitFootman, Faction: "Red"},
		"Brodo":  {Type: UnitFootman, Faction: "Red"},

		"Gron":  {Type: UnitFootman, Faction: "White"},
		"Gnade": {Type: UnitFootman, Faction: "Black"},

		"Firril": {Type: UnitFootman, Faction: "Black"},

		"Ovo":       {Type: UnitFootman, Faction: "Green"},
		"Mare Elle": {Type: UnitShip, Faction: "Green"},

		"Winde":      {Type: UnitFootman, Faction: "Green"},
		"Mare Gond":  {Type: UnitShip, Faction: "Green"},
		"Mare Ovond": {Type: UnitShip, Faction: "Green"},
		"Mare Unna":  {Type: UnitShip, Faction: "Black"},

		"Tusser": {Type: UnitFootman, Faction: "White"},
		"Tige":   {Type: UnitFootman, Faction: "Black"},

		"Tond": {Type: UnitFootman, Faction: "Green"},

		"Leil":   {Type: UnitFootman, Faction: "Red"},
		"Limbol": {Type: UnitFootman, Faction: "Green"},
		"Worp":   {Type: UnitFootman, Faction: "Yellow"},
	}

	orders := []Order{
		// Auto-success
		{Type: OrderMove, Origin: "Emman", Destination: "Erren"},

		// PvP battle with supported defender
		{Type: OrderMove, Origin: "Lomone", Destination: "Lusía"},
		{Type: OrderSupport, Origin: "Brodo", Destination: "Lusía"},

		// PvP battle, no defender
		{Type: OrderMove, Origin: "Gron", Destination: "Gewel"},
		{Type: OrderMove, Origin: "Gnade", Destination: "Gewel"},

		// PvE battle
		{Type: OrderMove, Origin: "Firril", Destination: "Furie"},

		// PvE battle, transport not attacked
		{Type: OrderMove, Origin: "Ovo", Destination: "Zona"},
		{Type: OrderTransport, Origin: "Mare Elle"},

		// PvE battle, transport attacked
		{Type: OrderMove, Origin: "Winde", Destination: "Fond"},
		{Type: OrderTransport, Origin: "Mare Gond"},
		{Type: OrderTransport, Origin: "Mare Ovond"},
		{Type: OrderMove, Origin: "Mare Unna", Destination: "Mare Ovond"},

		// Border battle
		{Type: OrderMove, Origin: "Tusser", Destination: "Tige"},
		{Type: OrderMove, Origin: "Tige", Destination: "Tusser"},

		// Danger zone, dependent move
		{Type: OrderMove, Origin: "Tond", Destination: "Tige"},

		// Move cycle
		{Type: OrderMove, Origin: "Leil", Destination: "Limbol"},
		{Type: OrderMove, Origin: "Limbol", Destination: "Worp"},
		{Type: OrderMove, Origin: "Worp", Destination: "Leil"},
	}

	game := newMockGame()
	placeUnits(units, game.board)
	placeOrders(orders, game.board)

	return game, orders
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

func (MockMessenger) SendWinner(winner PlayerFaction) error {
	return nil
}

func (MockMessenger) ClearMessages() {}
