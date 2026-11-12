package login

import (
	"context"

	"github.com/xoctopus/httpx/example/pkg/modules/user"
	"github.com/xoctopus/httpx/pkg/httpx"
)

type UserRegister struct {
	httpx.MethodPost `path:"/register" summary:"用户注册"`

	Body user.RegisterParam `in:"body"`
}

func (r *UserRegister) Output(ctx context.Context) (any, error) {
	return user.Register(ctx, &r.Body)
}
