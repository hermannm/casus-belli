package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
)

func getQueryParam(query url.Values, paramName string) (string, error) {
	value := query.Get(paramName)
	if value == "" {
		return "", fmt.Errorf("required query param '%s' was blank", paramName)
	}

	unescaped, err := url.QueryUnescape(value)
	if err != nil {
		return "", wrap.Errorf(err, "failed to parse query param '%s'", paramName)
	}

	return unescaped, nil
}

func sendJSON(res http.ResponseWriter, value any) {
	res.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(res).Encode(value); err != nil {
		err = wrap.Error(err, "failed to serialize response")
		sendServerError(res, err)
		log.Error(err)
	}
}

func sendClientError(res http.ResponseWriter, err error) {
	http.Error(res, err.Error(), http.StatusBadRequest)
}

func sendServerError(res http.ResponseWriter, err error) {
	http.Error(res, err.Error(), http.StatusInternalServerError)
}

// .NET ClientWebSockets, which we use for the Godot game client, do not provide a way to get the
// HTTP response message from a failed WebSocket connect request. Therefore, we have to pass the
// error message through a response header instead.
//
// See https://github.com/dotnet/runtime/issues/19405
func sendClientErrorWithHeader(res http.ResponseWriter, err error) {
	errMessage := err.Error()
	res.Header().Set("Error", errMessage)
	http.Error(res, errMessage, http.StatusBadRequest)
}

func sendServerErrorWithHeader(res http.ResponseWriter, err error) {
	errMessage := err.Error()
	res.Header().Set("Error", errMessage)
	http.Error(res, errMessage, http.StatusInternalServerError)
}
