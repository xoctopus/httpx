package login

import (
	"context"

	"github.com/xoctopus/httpx/internal/cmd/example/modules/user"
	"github.com/xoctopus/httpx/pkg/httpx"
)

type LoginByUsername struct {
	httpx.MethodPut `path:"/login" summary:"用户名密码登陆"`

	user.LoginParam `in:"body" mime:"json"`
}

func (r *LoginByUsername) Output(ctx context.Context) (any, error) {
	return user.LoginValidateByPassword(ctx, &r.LoginParam)
}
