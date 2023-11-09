package api

import (
	"encoding/json"
	"errors"
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
		sendServerError(res, err, "")
		log.Error(err, "")
	}
}

func sendClientError(res http.ResponseWriter, err error, message string) {
	sendError(res, err, message, http.StatusBadRequest)
}

func sendServerError(res http.ResponseWriter, err error, message string) {
	sendError(res, err, message, http.StatusInternalServerError)
}

func sendError(res http.ResponseWriter, err error, message string, statusCode int) {
	if err == nil {
		err = errors.New(message)
	} else if message != "" {
		err = wrap.Error(err, message)
	}

	http.Error(res, err.Error(), statusCode)
}
