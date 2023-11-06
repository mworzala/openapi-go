package oapi_rt

import "net/http"

type Middleware interface {
	Run(next http.Handler) http.Handler
}
