package main

import (
	"github.com/immerse-ntnu/hermannia/server/api"
	"github.com/immerse-ntnu/hermannia/server/app"
)

func main() {
	api.StartAPI(":7000", true, app.Games)
}
