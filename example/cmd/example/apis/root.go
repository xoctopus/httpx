package apis

import (
	"context"

	"github.com/xoctopus/confx/pkg/confjwt"
	"github.com/xoctopus/httpx/example/cmd/example/apis/login"
	"github.com/xoctopus/httpx/example/cmd/example/apis/providers"
	"github.com/xoctopus/httpx/example/cmd/example/apis/user"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/routex"
)

var (
	Root = httpx.GroupRouter("/example")
	v0   = httpx.GroupRouter("/v0")
	auth = httpx.NewRouter(&confjwt.Auth{}, &providers.CurrentUserInfo{})
)

type HealthCheck struct {
	httpx.MethodGet `path:"/health-check"`
}

func (HealthCheck) Output(ctx context.Context) (any, error) {
	return nil, nil
}

func init() {
	auth = auth.With(
		httpx.GroupRouter("/user").With(user.R),
	)

	v0 = v0.With(
		auth,
		httpx.NewRouter(&routex.OpenAPI{}),
		httpx.NewRouter(&HealthCheck{}),
		httpx.NewRouter(&login.LoginByUsername{}),
		httpx.NewRouter(&login.UserRegister{}),
		httpx.NewRouter(&login.CookieRefresh{}),
	)

	Root = Root.With(v0)
}

// TODO
// examples:
// for preprocess operator. [middlewares for validation]
// for postponed operator. [middlewares for logging]
// for serving openapi with upgrader.Upgrader
