package apis

import (
	"github.com/xoctopus/httpx/example/apis/org"
	"github.com/xoctopus/httpx/example/apis/user"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/routers"
)

var R = httpx.GroupRouter("/api/example/").With(
	httpx.NewRouter(&routers.OpenAPI{}),
	httpx.GroupRouter("v0").With(
		user.R,
		org.R,
	),
)
