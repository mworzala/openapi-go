// Code generated with openapi-go DO NOT EDIT.
package v1

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	oapi_rt "github.com/mworzala/openapi-go/pkg/oapi-rt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type TerraformServer interface {
	GetPlayerSession(ctx context.Context, playerId string) ([]byte, error)
	GetLocalSession(ctx context.Context, playerId string, worldId string) ([]byte, error)
}

type TerraformServerWrapper struct {
	log         *zap.SugaredLogger
	middlewares []oapi_rt.Middleware
	handler     TerraformServer
}

type TerraformServerWrapperParams struct {
	fx.In

	Log     *zap.SugaredLogger
	Handler TerraformServer
}

func NewTerraformServerWrapper(p TerraformServerWrapperParams) (*TerraformServerWrapper, error) {
	sw := &TerraformServerWrapper{
		log:     p.Log.With("handler", "terraform (wrapper)"),
		handler: p.Handler,
	}

	return sw, nil
}

func (sw *TerraformServerWrapper) Apply(r chi.Router) {
	r.Route("/v1/internal", func(r chi.Router) {

		r.Get("/terraform/session/{playerId}", sw.GetPlayerSession)
		r.Get("/terraform/session/{playerId}/{worldId}", sw.GetLocalSession)
	})
}

func (sw *TerraformServerWrapper) GetPlayerSession(w http.ResponseWriter, r *http.Request) {
	// Validate Parameters

	// Read Parameters
	playerId := chi.URLParam(r, "playerId")

	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code200, err := sw.handler.GetPlayerSession(r.Context(), playerId)
		if err != nil {
			oapi_rt.WriteGenericError(w, err)
			return
		}

		if code200 != nil {
			w.Header().Set("content-type", "application/vnd.terraform.player_session")
			w.WriteHeader(200)
			_, _ = w.Write(code200)
			return
		}

		w.WriteHeader(404)
	})
	for _, middleware := range sw.middlewares {
		handler = middleware.Run(handler)
	}
	handler.ServeHTTP(w, r)
}

func (sw *TerraformServerWrapper) GetLocalSession(w http.ResponseWriter, r *http.Request) {
	// Validate Parameters

	// Read Parameters
	playerId := chi.URLParam(r, "playerId")
	worldId := chi.URLParam(r, "worldId")

	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code200, err := sw.handler.GetLocalSession(r.Context(), playerId, worldId)
		if err != nil {
			oapi_rt.WriteGenericError(w, err)
			return
		}

		if code200 != nil {
			w.Header().Set("content-type", "application/vnd.terraform.local_session")
			w.WriteHeader(200)
			_, _ = w.Write(code200)
			return
		}

		w.WriteHeader(404)
	})
	for _, middleware := range sw.middlewares {
		handler = middleware.Run(handler)
	}
	handler.ServeHTTP(w, r)
}
