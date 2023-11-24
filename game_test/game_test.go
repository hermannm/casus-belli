package game_test

import (
	"testing"

	"hermannm.dev/bfh-server/game"
)

func TestNonWinterOrders(t *testing.T) {
	testCases := []struct {
		name     string
		units    unitMap
		control  controlMap
		orders   []game.Order
		expected expectedUnits
	}{
		{
			name: "UncontestedMove",
			units: unitMap{
				"Emman": {Type: game.UnitFootman, Faction: white},
			},
			orders: []game.Order{
				{Type: game.OrderMove, Origin: "Emman", Destination: "Erren"},
			},
			expected: expectedUnits{
				"Erren": "Emman",
				"Emman": "",
			},
		},
		{
			name: "SingleplayerBattle",
			units: unitMap{
				"Furie": {Type: game.UnitHorse, Faction: black},
			},
			orders: []game.Order{
				{Type: game.OrderMove, Origin: "Furie", Destination: "Firril"},
			},
			expected: expectedUnits{
				"Furie":  "Furie",
				"Firril": "",
			},
		},
		{
			name: "SingleplayerBattleWithSupport",
			units: unitMap{
				"Furie":      {Type: game.UnitHorse, Faction: black},
				"Mare Ovond": {Type: game.UnitShip, Faction: black},
			},
			orders: []game.Order{
				{Type: game.OrderMove, Origin: "Furie", Destination: "Firril"},
				{Type: game.OrderSupport, Origin: "Mare Ovond", Destination: "Firril"},
			},
			expected: expectedUnits{
				"Furie":  "",
				"Firril": "Furie",
			},
		},
		{
			name: "MultiplayerBattleNoDefender",
			units: unitMap{
				"Gron":  {Type: game.UnitFootman, Faction: white},
				"Gewel": {Type: game.UnitHorse, Faction: black},
			},
			control: controlMap{
				"Gnade": black,
			},
			orders: []game.Order{
				{Type: game.OrderMove, Origin: "Gron", Destination: "Gnade"},
				{Type: game.OrderMove, Origin: "Gewel", Destination: "Gnade"},
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
				"Lomone": {Type: game.UnitFootman, Faction: green},
				"Lusía":  {Type: game.UnitFootman, Faction: red},
				"Brodo":  {Type: game.UnitFootman, Faction: red},
			},
			orders: []game.Order{
				{Type: game.OrderMove, Origin: "Lomone", Destination: "Lusía"},
				{Type: game.OrderSupport, Origin: "Brodo", Destination: "Lusía"},
			},
			expected: expectedUnits{
				"Lusía":  "Lusía",
				"Lomone": "",
			},
		},
		{
			name: "BorderBattle",
			units: unitMap{
				"Tusser": {Type: game.UnitFootman, Faction: white},
				"Tige":   {Type: game.UnitHorse, Faction: black},
			},
			orders: []game.Order{
				{Type: game.OrderMove, Origin: "Tusser", Destination: "Tige"},
				{Type: game.OrderMove, Origin: "Tige", Destination: "Tusser"},
			},
			expected: expectedUnits{
				"Tige":   "Tusser",
				"Tusser": "",
			},
		},
		{
			name: "UncontestedHorseMove",
			units: unitMap{
				"Lomone": {Type: game.UnitHorse, Faction: red},
			},
			control: controlMap{
				"Limbol": red,
				"Worp":   green,
			},
			orders: []game.Order{
				{
					Type:              game.OrderMove,
					Origin:            "Lomone",
					Destination:       "Limbol",
					SecondDestination: "Worp",
				},
			},
			expected: expectedUnits{
				"Worp":   "Lomone",
				"Limbol": "",
				"Lomone": "",
			},
		},
		{
			name: "HorseMovesCuttingSupport",
			units: unitMap{
				"Worp":      {Type: game.UnitHorse, Faction: green},
				"Winde":     {Type: game.UnitHorse, Faction: green},
				"Lomone":    {Type: game.UnitHorse, Faction: red},
				"Lusía":     {Type: game.UnitHorse, Faction: red},
				"Mare Illa": {Type: game.UnitShip, Faction: green},
				"Mare Duna": {Type: game.UnitShip, Faction: green},
				"Morone":    {Type: game.UnitFootman, Faction: green},
				"Brodo":     {Type: game.UnitFootman, Faction: green},
			},
			control: controlMap{
				"Limbol": red,
				"Leil":   red,
			},
			orders: []game.Order{
				{
					Type:              game.OrderMove,
					Origin:            "Worp",
					Destination:       "Limbol",
					SecondDestination: "Lomone",
				},
				{
					Type:              game.OrderMove,
					Origin:            "Winde",
					Destination:       "Leil",
					SecondDestination: "Lusía",
				},
				{Type: game.OrderSupport, Origin: "Lusía", Destination: "Lomone"},
				{Type: game.OrderSupport, Origin: "Lomone", Destination: "Lusía"},
				{Type: game.OrderSupport, Origin: "Mare Illa", Destination: "Lomone"},
				{Type: game.OrderSupport, Origin: "Mare Duna", Destination: "Lomone"},
				{Type: game.OrderSupport, Origin: "Morone", Destination: "Lusía"},
				{Type: game.OrderSupport, Origin: "Brodo", Destination: "Lusía"},
			},
			expected: expectedUnits{
				"Lomone": "Worp",
				"Lusía":  "Winde",
				"Limbol": "",
				"Leil":   "",
				"Worp":   "",
				"Winde":  "",
			},
		},
		{
			name: "Transport",
			units: unitMap{
				"Ovo":       {Type: game.UnitFootman, Faction: green},
				"Mare Elle": {Type: game.UnitShip, Faction: green},
			},
			control: controlMap{
				"Zona": white,
			},
			orders: []game.Order{
				{Type: game.OrderMove, Origin: "Ovo", Destination: "Zona"},
				{Type: game.OrderTransport, Origin: "Mare Elle"},
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
				"Winde":      {Type: game.UnitFootman, Faction: green},
				"Mare Gond":  {Type: game.UnitShip, Faction: green},
				"Mare Ovond": {Type: game.UnitShip, Faction: green},
				"Mare Unna":  {Type: game.UnitShip, Faction: black},
			},
			control: controlMap{
				"Fond": black,
			},
			orders: []game.Order{
				{Type: game.OrderMove, Origin: "Winde", Destination: "Fond"},
				{Type: game.OrderTransport, Origin: "Mare Gond"},
				{Type: game.OrderTransport, Origin: "Mare Ovond"},
				{Type: game.OrderMove, Origin: "Mare Unna", Destination: "Mare Ovond"},
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
				"Leil":   {Type: game.UnitFootman, Faction: red},
				"Limbol": {Type: game.UnitFootman, Faction: green},
				"Worp":   {Type: game.UnitFootman, Faction: yellow},
			},
			orders: []game.Order{
				{Type: game.OrderMove, Origin: "Leil", Destination: "Limbol"},
				{Type: game.OrderMove, Origin: "Limbol", Destination: "Worp"},
				{Type: game.OrderMove, Origin: "Worp", Destination: "Leil"},
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
			game, board := newMockGame(
				t,
				testCase.units,
				testCase.control,
				testCase.orders,
				game.SeasonSpring,
			)
			game.ResolveNonWinterOrders(testCase.orders)
			testCase.expected.check(t, board, testCase.units)
		})
	}
}
