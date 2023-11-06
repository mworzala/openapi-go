// Code generated with openapi-go DO NOT EDIT.
package v1

import (
	"github.com/go-chi/chi/v5"
	oapi_rt "github.com/mworzala/openapi-go/pkg/oapi-rt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PublicServer interface {
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
}

func NewPublicServerWrapper(p PublicServerWrapperParams) (*PublicServerWrapper, error) {
	sw := &PublicServerWrapper{
		log:     p.Log.With("handler", "public (wrapper)"),
		handler: p.Handler,
	}

	return sw, nil
}

func (sw *PublicServerWrapper) Apply(r chi.Router) {
	r.Route("/v1/internal", func(r chi.Router) {

	})
}
