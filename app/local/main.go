package main

import (
	"net/http"

	"github.com/immerse-ntnu/hermannia/server/api"
)

func main() {
	api.RegisterEndpoints(nil)
	http.ListenAndServe(":7000", nil)
}
