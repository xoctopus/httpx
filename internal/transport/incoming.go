package transport

import (
	"context"
	"net/http"

	"github.com/xoctopus/logx"

	"github.com/xoctopus/httpx/internal/payload/transformer"
	"github.com/xoctopus/httpx/internal/request"
	"github.com/xoctopus/httpx/internal/response"
)

type Incoming interface {
	UnmarshalOperator(ctx context.Context, r request.Request, op any) error
	WriteResponse(ctx context.Context, rw http.ResponseWriter, result any, r request.Request)
}

func NewIncoming(_ context.Context, _ any) (Incoming, error) {
	return &incoming{}, nil
}

type incoming struct{}

func (i *incoming) UnmarshalOperator(_ context.Context, r request.Request, op any) error {
	return transformer.UnmarshalRequest(r, op)
}

func (i *incoming) WriteResponse(ctx context.Context, rw http.ResponseWriter, result any, r request.Request) {
	var w response.Writer

	defer func() {
		if w != nil {
			if err := w.WriteResponse(ctx, rw, r); err != nil {
				logx.From(ctx).Error(err)
			}
		}
	}()

	if upgrader, ok := result.(Upgrader); ok {
		if err := upgrader.Upgrade(rw, r.Underlying()); err != nil {
			w = response.ErrorResponse(err).(response.Writer)
		}
		return
	}

	switch v := result.(type) {
	case error:
		w = response.ErrorResponse(v).(response.Writer)
	default:
		w = response.New(v).(response.Writer)
	}
}
