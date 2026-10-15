package httpx

import (
	"net/http"
	"slices"

	"github.com/xoctopus/httpx/internal/operator"
	"github.com/xoctopus/httpx/internal/route"
)

type (
	Route  = route.Route
	Routes = route.Routes
	Router = route.Router

	Middleware = route.Middleware
	Handler    = route.Handler
)

func NewRouter(operators ...Operator) Router {
	return route.NewRouter(operators...)
}

func GroupRouter(group string) Router {
	return NewRouter(operator.GroupOperator(group))
}

func BasePathRouter(path string) Router {
	return NewRouter(operator.BasePathOperator(path))
}

func ApplyMiddlewares(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		last := final
		for _, middleware := range slices.Backward(middlewares) {
			last = middleware(last)
		}
		return last
	}
}
