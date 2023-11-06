package oapi_rt

import (
	"errors"
	"net/http"
)

type ErrorHandler interface {
	HandleUnknownError(w http.ResponseWriter, r *http.Request, err error)
}

type WritableError interface {
	error

	Write(w http.ResponseWriter, r *http.Request)
}

func WriteMissingParamError(w http.ResponseWriter, param string) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte("missing required parameter: " + param))
}

func WriteGenericError(w http.ResponseWriter, err error) {
	var werr WritableError
	if errors.As(err, &werr) {
		werr.Write(w, nil)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}
