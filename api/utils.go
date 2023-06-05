package api

import (
	"net/http"
	"net/url"
)

func checkParams(req *http.Request, keys ...string) (params url.Values, ok bool) {
	params = req.URL.Query()

	for _, key := range keys {
		if params.Get(key) == "" {
			return nil, false
		}
	}

	return params, true
}
