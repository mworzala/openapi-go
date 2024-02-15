// Code generated with openapi-go DO NOT EDIT.
package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	oapi_rt "github.com/mworzala/openapi-go/pkg/oapi-rt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PublicServer interface {
	GetTestPlainResp(ctx context.Context) (string, error)
	GetMapWorld(ctx context.Context, id string, abc int, accept string, req *MapManualTriggerWebhook) (*GetMapWorldResponse, *MapManualTriggerWebhook, error)
}

type PublicServerWrapper struct {
	log         *zap.SugaredLogger
	middlewares []oapi_rt.Middleware
	handler     PublicServer
}

type PublicServerWrapperParams struct {
	fx.In
	Log     *zap.SugaredLogger
	Handler PublicServer

	Middleware []oapi_rt.Middleware `group:"public_middleware"`
}

func NewPublicServerWrapper(p PublicServerWrapperParams) (*PublicServerWrapper, error) {
	sw := &PublicServerWrapper{
		log:         p.Log.With("handler", "public (wrapper)"),
		handler:     p.Handler,
		middlewares: p.Middleware,
	}

	return sw, nil
}

func (sw *PublicServerWrapper) Apply(r chi.Router) {
	r.Route("/v1/public", func(r chi.Router) {
		r.Get("/test/plain_resp", sw.GetTestPlainResp)
		r.Get("/maps/{id}/world", sw.GetMapWorld)
	})
}

func (sw *PublicServerWrapper) GetTestPlainResp(w http.ResponseWriter, r *http.Request) {
	var err error
	_ = err // Sometimes we don't use it but need that not to be an error

	// Read Parameters

	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := oapi_rt.NewContext(r.Context(), r)

		code200, err := sw.handler.GetTestPlainResp(ctx)
		if err != nil {
			oapi_rt.WriteGenericError(w, err)
			return
		}

		if code200 != "" {
			w.Header().Set("content-type", "text/plain")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(code200))

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

func (sw *PublicServerWrapper) GetMapWorld(w http.ResponseWriter, r *http.Request) {
	var err error
	_ = err // Sometimes we don't use it but need that not to be an error

	// Read Parameters

	abcStr := r.URL.Query().Get("abc")
	var abc int
	abc, err = strconv.Atoi(abcStr)
	if err != nil {
		oapi_rt.WriteGenericError(w, err)
		return
	}

	accept := r.Header.Get("accept")

	id := chi.URLParam(r, "id")

	// Read Body
	var body MapManualTriggerWebhook
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		oapi_rt.WriteGenericError(w, err)
		return
	}

	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := oapi_rt.NewContext(r.Context(), r)

		code200, code201, err := sw.handler.GetMapWorld(ctx, id, abc, accept, &body)
		if err != nil {
			oapi_rt.WriteGenericError(w, err)
			return
		}

		if code200 != nil {
			switch {
			case code200.Polar != nil:
				w.Header().Set("content-type", "application/vnd.hollowcube.polar")
				w.WriteHeader(200)
				_, _ = w.Write(code200.Polar)
				return
			case code200.Anvil != nil:
				w.Header().Set("content-type", "application/vnd.hollowcube.anvil")
				w.WriteHeader(200)
				_, _ = w.Write(code200.Anvil)
				return
			}
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

		w.WriteHeader(204)
	})
	for _, middleware := range sw.middlewares {
		handler = middleware.Run(handler)
	}
	handler.ServeHTTP(w, r)
}
