package game

import (
	"math/rand"
	"time"
)

func RollDice() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(6) + 1
}
