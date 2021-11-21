package utils

import (
	t "immerse-ntnu/hermannia/server/types"
)

func Sailable(area t.BoardArea) bool {
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
