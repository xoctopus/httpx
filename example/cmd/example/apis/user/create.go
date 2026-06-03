package user

import (
	"context"

	"github.com/xoctopus/httpx/example/pkg/modules/user"
	"github.com/xoctopus/httpx/pkg/httpx"
)

type RegisterMember struct {
	httpx.MethodPost `path:"/register"`

	Body user.RegisterReq `in:"body"`
}

func (req *RegisterMember) Output(ctx context.Context) (any, error) {
	return httpx.WrapResponse[*user.RegisterRsp](
		&user.RegisterRsp{
			UserID:   100,
			Username: req.Body.Username,
		},
		httpx.WithStatusCode(httpx.STATUS__OK),
	), nil
}
