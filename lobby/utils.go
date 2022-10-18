package lobby

import (
	"net/http"
	"net/url"
)

// Checks the given request for the existence of the provided parameter keys.
// If all exist, returns the parameters, otherwise returns ok = false.
func checkParams(req *http.Request, keys ...string) (params url.Values, ok bool) {
	params = req.URL.Query()

	for _, key := range keys {
		if params.Get(key) == "" {
			return nil, false
		}
	}

	return params, true
}
