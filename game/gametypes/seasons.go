package gametypes

// Current season of a round. Affects the type of orders that can be played.
type Season string

// Rounds where only build and internal move orders are allowed.
const SeasonWinter Season = "winter"

// Rounds where only move, support, transport and besiege orders are allowed.
const (
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonFall   Season = "fall"
)

func (season Season) Next() Season {
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
