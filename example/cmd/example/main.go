package main

import (
	"context"

	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/httpx/example/cmd/example/apis"
	"github.com/xoctopus/httpx/example/pkg/modules/user"
	"github.com/xoctopus/httpx/pkg/confhttp"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/middlex"
	"github.com/xoctopus/httpx/pkg/routex"
)

func main() {
	ctx := logx.With(context.Background(), logx.NewStd())

	h, err := routex.NewHandler(apis.R, "example")
	if err != nil {
		panic(err)
	}

	// services injections
	carriers := []contextx.Carrier{
		user.Carry(user.New()),
	}

	h = httpx.ApplyMiddlewares(
		middlex.InjectContext(carriers...),
	)(h)

	if err = confhttp.ListenAndServe(ctx, "0.0.0.0:9001", h); err != nil {
		panic(err)
	}
}
