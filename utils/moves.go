package utils

import (
	"immerse/hermannia/server/types"
)

func Sailable(area types.BoardArea) bool {
	if area.Sea {
		return true
	}

	for _, neighbor := range area.Neighbors {
		if neighbor.Area.Sea {
			return true
		}
	}

	return false
}
