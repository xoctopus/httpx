package user

import (
	"context"
	"net/http"
	"time"

	"github.com/xoctopus/httpx/pkg/httpx"
)

type Cookie struct {
	httpx.MethodPost `path:"/cookie-ping-pong"`
	Token            string `in:"cookie" name:"token,omitempty"`
}

func (req *Cookie) Output(ctx context.Context) (any, error) {
	return httpx.WrapResponse[any](
		nil,
		httpx.WithCookies(&http.Cookie{
			Name:    "token",
			Value:   req.Token,
			Expires: time.Now().Add(24 * time.Hour),
		}),
		httpx.WithStatusCode(http.StatusOK),
	), nil
}
