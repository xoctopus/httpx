package httpx

import (
	"github.com/xoctopus/httpx/internal/operator"
	"github.com/xoctopus/httpx/internal/route"
)

type (
	Route  = route.Route
	Routes = route.Routes
	Router = route.Router
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
