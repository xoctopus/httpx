package user

import (
	"context"
	"net/http"
	"time"

	"github.com/xoctopus/httpx/pkg/httpx"
)

type CookieRefresh struct {
	httpx.MethodPut `path:"/cookie-refresh"`
	Token           string `in:"cookie" name:"token,omitempty"`
}

func (req *CookieRefresh) Output(ctx context.Context) (any, error) {
	return httpx.WrapResponse[any](
		nil,
		httpx.WithCookies(&http.Cookie{
			Name:    "token",
			Value:   "my-token",
			Expires: time.Now().Add(24 * time.Hour),
		}),
		httpx.WithStatusCode(httpx.STATUS__NO_CONTENT),
	), nil
}
