// Code generated with openapi-go DO NOT EDIT.
package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	oapi_rt "github.com/mworzala/openapi-go/pkg/oapi-rt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type InternalServer interface {
	CreateSession(ctx context.Context, playerId string, req *CreateSessionRequest) (interface{}, error)
	DeleteSession(ctx context.Context, playerId string) error
}

type InternalServerWrapper struct {
	log         *zap.SugaredLogger
	middlewares []oapi_rt.Middleware
	handler     InternalServer
}

type InternalServerWrapperParams struct {
	fx.In

	Log     *zap.SugaredLogger
	Handler InternalServer
}

func NewInternalServerWrapper(p InternalServerWrapperParams) (*InternalServerWrapper, error) {
	sw := &InternalServerWrapper{
		log:     p.Log.With("handler", "internal (wrapper)"),
		handler: p.Handler,
	}

	return sw, nil
}

func (sw *InternalServerWrapper) Apply(r chi.Router) {
	r.Route("/v1/internal", func(r chi.Router) {

		r.Post("/session/{playerId}", sw.CreateSession)
		r.Delete("/session/{playerId}", sw.DeleteSession)
	})
}

func (sw *InternalServerWrapper) CreateSession(w http.ResponseWriter, r *http.Request) {
	// Validate Parameters

	// Read Parameters
	playerId := chi.URLParam(r, "playerId")

	// Read Body
	var body CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		oapi_rt.WriteGenericError(w, err)
		return
	}

	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code201, err := sw.handler.CreateSession(r.Context(), playerId, &body)
		if err != nil {
			oapi_rt.WriteGenericError(w, err)
			return
		}

		if code201 != nil {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(201)
			if err = json.NewEncoder(w).Encode(code201); err != nil {
				sw.log.Errorw("failed to encode response", "err", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		// !! UNDEFINED EMPTY BEHAVIOR !!
		// Set `x-type: empty` on a response to define this behavior.
		sw.log.Errorw("empty response")
		w.WriteHeader(http.StatusInternalServerError)
	})
	for _, middleware := range sw.middlewares {
		handler = middleware.Run(handler)
	}
	handler.ServeHTTP(w, r)
}

func (sw *InternalServerWrapper) DeleteSession(w http.ResponseWriter, r *http.Request) {
	// Validate Parameters

	// Read Parameters
	playerId := chi.URLParam(r, "playerId")

	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := sw.handler.DeleteSession(r.Context(), playerId)
		if err != nil {
			oapi_rt.WriteGenericError(w, err)
			return
		}

		w.WriteHeader(200)
	})
	for _, middleware := range sw.middlewares {
		handler = middleware.Run(handler)
	}
	handler.ServeHTTP(w, r)
}
