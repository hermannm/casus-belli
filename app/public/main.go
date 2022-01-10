package main

import (
	"net/http"

	"github.com/immerse-ntnu/hermannia/server/api"
	"github.com/immerse-ntnu/hermannia/server/app"
)

func main() {
	api.RegisterEndpoints(nil)
	api.RegisterLobbyCreationEndpoints(nil, app.Games)
	http.ListenAndServe(":7000", nil)
}
