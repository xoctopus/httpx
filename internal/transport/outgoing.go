package transport

import (
	"context"
	"net/http"
	"reflect"

	"github.com/xoctopus/httpx/internal/method"
	"github.com/xoctopus/httpx/internal/operator"
	"github.com/xoctopus/httpx/internal/payload/path"
	"github.com/xoctopus/httpx/internal/payload/transformer"
)

type Outgoing interface {
	NewRequest(ctx context.Context, v any) (*http.Request, error)
}

func NewRequest(ctx context.Context, v any) (*http.Request, error) {
	t, err := NewOutgoingTransport(ctx, v)
	if err != nil {
		return nil, err
	}

	return t.NewRequest(ctx, v)

}

func NewOutgoingTransport(ctx context.Context, r any) (Outgoing, error) {
	rt := reflect.TypeOf(r)
	for rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}

	o := &outgoing{typ: rt}

	if x, ok := r.(method.Describer); ok {
		o.method = x.Method()
	}

	if x, ok := r.(operator.PathDescriber); ok {
		o.path = x.Path()
	}

	if o.path == "" {
		if o.typ.Kind() == reflect.Struct {
			o.path, _ = path.ResolveFromTag(o.typ)
		}
	}

	return o, nil
}

type outgoing struct {
	method string
	path   string
	typ    reflect.Type
}

func (o outgoing) Method() string {
	return o.method
}

func (o outgoing) Path() string {
	return o.path
}

func (o outgoing) NewRequest(ctx context.Context, v any) (*http.Request, error) {
	return transformer.NewRequest(ctx, o.Method(), o.Path(), v)
}
