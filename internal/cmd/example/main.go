package main

import (
	"context"
	"time"

	"github.com/xoctopus/confx/pkg/confjwt"
	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/logx"

	"github.com/xoctopus/httpx/internal/cmd/example/apis"
	"github.com/xoctopus/httpx/pkg/confhttp"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/midx"
	"github.com/xoctopus/httpx/pkg/routex"
)

func main() {
	ctx := logx.With(context.Background(), logx.NewStd())

	h, err := routex.NewHandler(
		apis.Root,
		"example@v0.0.1",
		midx.DefaultLogHandler(),
	)
	if err != nil {
		panic(err)
	}

	// services injections
	carriers := []contextx.Carrier{
		confjwt.CarryJWT(&confjwt.JWT{
			Issuer:  "example",
			ExpIn:   types.Duration(time.Hour),
			SignKey: "x",
		}),
	}

	h = httpx.ApplyMiddlewares(
		midx.InjectContext(carriers...),
		midx.DefaultCORS(),
	)(h)

	if err = confhttp.ListenAndServe(ctx, "0.0.0.0:9001", h); err != nil {
		panic(err)
	}
}
