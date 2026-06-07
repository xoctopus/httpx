package apis

import (
	"context"

	"github.com/xoctopus/httpx/example/cmd/example/apis/user"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/routex"
)

var R = httpx.GroupRouter("/api/example/v0").With(
	httpx.NewRouter(&routex.OpenAPI{}),
	httpx.GroupRouter("/user").With(user.R),
	// httpx.GroupRouter("org").With(org.R),
	httpx.NewRouter(&HealthCheck{}),
)

type HealthCheck struct {
	httpx.MethodGet `path:"/health-check"`
}

func (HealthCheck) Output(ctx context.Context) (any, error) {
	return nil, nil
}

// TODO
// examples:
// for preprocess operator. [middlewares for validation]
// for postponed operator. [middlewares for logging]
// for serving openapi with upgrader.Upgrader
