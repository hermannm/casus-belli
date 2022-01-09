package main

import (
	"github.com/immerse-ntnu/hermannia/server/api"
	"github.com/immerse-ntnu/hermannia/server/app"
)

func main() {
	api.StartPublic(":7000", app.Games)
}
