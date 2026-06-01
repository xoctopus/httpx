package route

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/httpx/internal/operator"
	"github.com/xoctopus/httpx/internal/payload/path"
	"github.com/xoctopus/httpx/internal/request"
	"github.com/xoctopus/httpx/internal/transport"
	"github.com/xoctopus/httpx/internal/types"
)

func ApplyMiddlewares(mw ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		last := final
		for i := len(mw) - 1; i >= 0; i-- {
			last = mw[i](last)
		}
		return last
	}
}

type Middleware = func(http.Handler) http.Handler

type Handler interface {
	http.Handler

	OperationID() string
	Method() string
	Path() string
	PathSegments() path.Segments
	Summary() string
	Description() string
	Deprecated() bool

	Operators() []*operator.Factory
}

func NewRouteHandlers(route Route, service string, middlewares ...Middleware) ([]Handler, error) {
	h := &handler{
		service:    service,
		middleware: ApplyMiddlewares(middlewares...),
		once:       &sync.Once{},
	}

	base, uri := "/", ""

	err := route.Range(func(f *operator.Factory, i int) error {
		m := NewMeta(f)

		if m.BasePath != "" {
			base = m.BasePath
		}

		if m.Path != "" {
			uri += "/" + m.Path
		}

		if f.IsLast {
			h.operation = f.Type.Name()
			h.deprecated = m.Deprecated
			h.summary = m.Summary
			h.description = m.Description
			if m.Method != "" {
				h.method = m.Method
			}
		}

		if f.NoOutput {
			return nil
		}

		tt, err := transport.NewIncoming(context.Background(), f.New())
		if err != nil {
			return err
		}

		h.operators = append(h.operators, f)
		h.transformers = append(h.transformers, tt)

		return nil
	})
	if err != nil {
		return nil, err
	}

	h.segments = path.ParseSegments(path.Normalize(base + uri))

	var (
		methods  = strings.Split(h.method, ",")
		handlers = make([]Handler, 0, len(methods))
	)
	for _, m := range methods {
		if m == "" {
			continue
		}
		if h.method == m {
			handlers = append(handlers, h)
		} else {
			handlers = append(handlers, h.clone(m))
		}
	}

	return handlers, nil
}

type handler struct {
	service      string
	operation    string
	method       string
	segments     path.Segments
	summary      string
	deprecated   bool
	description  string
	operators    []*operator.Factory
	transformers []transport.Incoming
	middleware   Middleware

	once *sync.Once
	hh   http.Handler
}

func (h *handler) OperationID() string {
	return h.operation
}

func (h *handler) Method() string {
	return h.method
}

func (h *handler) Path() string {
	return h.segments.ParamString()
}

func (h *handler) PathSegments() path.Segments {
	return h.segments
}

func (h *handler) Summary() string {
	if h.summary == "" {
		return h.OperationID()
	}
	return h.summary
}

func (h *handler) Description() string {
	return h.description
}

func (h *handler) Deprecated() bool {
	return h.deprecated
}

func (h *handler) Operators() []*operator.Factory {
	return h.operators
}

func (h handler) clone(method string) Handler {
	h.method, h.operation = method, fmt.Sprintf("%s_%s", method, h.operation)
	return new(h)
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.once.Do(func() {
		var hh http.Handler = &_handler{handler: h}

		if h.middleware != nil {
			hh = h.middleware(hh)
		}

		for _, o := range h.operators {
			if x, ok := o.Operator.(WithPreHandlerMiddleware); ok {
				hh = x.PreHandlerMiddleware(hh)
			}
		}

		h.hh = hh
	})
	h.hh.ServeHTTP(rw, r)
}

type _handler struct {
	*handler
}

func (h *_handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = types.WithRequest(r.Context(), r)
		ri  = request.From(r)
	)

	for i := range h.operators {
		var (
			f = h.operators[i]    // operator factory
			t = h.transformers[i] // transport
			o = f.New()           // operator
		)

		if err := t.UnmarshalOperator(ctx, ri, o); err != nil {
			t.WriteResponse(ctx, rw, err, ri)
			return
		}

		if x, ok := o.(operator.Initializer); ok {
			if err := x.Init(ctx); err != nil {
				t.WriteResponse(ctx, rw, err, ri)
				return
			}
		}

		v, err := o.Output(ctx)
		if err != nil {
			t.WriteResponse(ctx, rw, err, ri)
			return
		}

		if !f.IsLast {
			if x, ok := o.(contextx.Provider); ok {
				ctx = x.WithContext(ctx)
			} else {
				switch u := v.(type) {
				case contextx.Provider:
					ctx = u.WithContext(ctx)
				case context.Context:
					ctx = u
				default:
					if k := f.ContextKey; k != nil {
						ctx = contextx.WithValue(ctx, k, v)
					}
				}
			}
			continue
		}

		t.WriteResponse(ctx, rw, v, ri)
	}
}

type WithPreHandlerMiddleware interface {
	PreHandlerMiddleware(h http.Handler) http.Handler
}
