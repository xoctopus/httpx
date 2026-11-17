package login

import (
	"context"
	"net/http"

	"github.com/xoctopus/httpx/internal/cmd/example/modules/user"
	"github.com/xoctopus/httpx/pkg/httpx"
)

type CookieRefresh struct {
	httpx.MethodPut `path:"/token-refresh" summary:"刷新凭证"`
	Token           string `in:"cookie" name:"token,omitempty"`
}

func (req *CookieRefresh) Output(ctx context.Context) (any, error) {
	res, err := user.RefreshToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	return httpx.WrapResponse[any](
		nil,
		httpx.WithResponseCookies(&http.Cookie{
			Name:    "token",
			Value:   res.Token,
			Expires: res.ExpiredAt,
		}),
		httpx.WithResponseStatusCode(httpx.STATUS__NO_CONTENT),
	), nil
}
