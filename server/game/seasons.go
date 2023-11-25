package game

import "hermannm.dev/enumnames"

// Current season of a round. Affects the type of orders that can be played.
type Season uint8

const (
	SeasonWinter Season = iota + 1
	SeasonSpring
	SeasonSummer
	SeasonFall
)

var seasonNames = enumnames.NewMap(map[Season]string{
	SeasonWinter: "Winter",
	SeasonSpring: "Spring",
	SeasonSummer: "Summer",
	SeasonFall:   "Fall",
})

func (season Season) String() string {
	return seasonNames.GetNameOrFallback(season, "INVALID")
}

func (season Season) next() Season {
	switch season {
	case SeasonWinter:
		return SeasonSpring
	case SeasonSpring:
		return SeasonSummer
	case SeasonSummer:
		return SeasonFall
	case SeasonFall:
		return SeasonWinter
	default:
		return SeasonWinter
	}
}
