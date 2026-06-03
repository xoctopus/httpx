package routex

import (
	"net/http"

	"github.com/xoctopus/httpx/internal/route"
	"github.com/xoctopus/httpx/pkg/httpx"
)

func NewHandler(r httpx.Router, service string, middleware ...httpx.Middleware) (http.Handler, error) {
	m, err := newmux(r, service, middleware...)
	if err != nil {
		return nil, err
	}

	h, err := m.Handler()
	if err != nil {
		return nil, err
	}
	return route.MethodOverwrite()(h), nil
}
