package user

import (
	"github.com/xoctopus/httpx/internal/cmd/example/apis/providers"
	"github.com/xoctopus/httpx/pkg/httpx"
)

var R = httpx.NewRouter()

func init() {
	R.Register(httpx.NewRouter(&ListUsers{}))
	R.Register(httpx.NewRouter(&FetchByUsername{}))
	R.Register(httpx.NewRouter(&providers.UserPermission{}, &DeleteUserByUserID{}))
	R.Register(httpx.NewRouter(&providers.UserPermission{}, &PromoteUserLevel{}))
}
