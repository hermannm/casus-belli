package gametypes

// Current season of a round (affects the type of orders that can be played).
// See Season constants for possible values.
type Season string

// Rounds where only build and internal move orders are allowed.
const SeasonWinter Season = "winter"

// Rounds where only move, support, transport and besiege orders are allowed.
const (
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonFall   Season = "fall"
)

// Returns the next season given the current season.
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
