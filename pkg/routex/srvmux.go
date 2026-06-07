package routex

import (
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/httpx/internal/openapi"
	"github.com/xoctopus/httpx/internal/operator"
	"github.com/xoctopus/httpx/internal/payload/path"
	"github.com/xoctopus/httpx/internal/route"
	"github.com/xoctopus/httpx/internal/types"
	oas "github.com/xoctopus/httpx/pkg/spec/openapi"
)

func newmux(root route.Router, name string, middlewares ...route.Middleware) (*mux, error) {
	customspec := false
	for _, r := range root.Routes() {
		if customspec {
			break
		}
		_ = r.Range(func(f *operator.Factory, i int) error {
			if f.IsLast {
				if _, ok := f.Operator.(*OpenAPI); ok {
					customspec = true
				}
			}
			return nil
		})
	}
	if !customspec {
		root.Register(route.NewRouter(&OpenAPI{}))
		// root.Register(route.NewRouter(&OpenAPIView{}))
	}

	spec := openapi.DefaultBuildFunc(root)
	spec.Title = name

	routes := root.Routes()
	handlers := make([]route.Handler, 0, len(routes))

	for i := range routes {
		// middleware for each route
		rh, err := route.NewHandlers(routes[i], name, middlewares...)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, rh...)
	}

	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].Path() < handlers[j].Path()
	})

	m := &mux{operations: &operations{spec: spec}}

	parts := strings.Split(name, "@")
	m.meta.Name = parts[0]
	if len(parts) >= 2 {
		m.meta.Version = parts[1]
	}

	for i := range handlers {
		m.register(handlers[i])
	}

	return m, nil
}

type mux struct {
	meta       types.ServerMeta
	operations *operations
	tree       *path.Tree[route.Handler]
	printer    *printer
}

func (m *mux) register(h route.Handler) {
	if m.tree == nil {
		m.tree = &path.Tree[route.Handler]{}
	}
	m.tree.Add(h)
}

func (m *mux) group() *group {
	g := &group{}

	if m.tree == nil {
		return g
	}

	for h := range m.tree.Route() {
		g.add(h, slices.Collect(h.PathSegments().Chunk())...)
	}

	// fmt.Println(g.String())
	return g
}

func (m *mux) add(r *http.ServeMux, h route.Handler) {
	method := h.Method()
	if method == "" {
		return
	}

	info := &types.OperationMeta{
		ServerMeta: m.meta,
		RequestMeta: types.RequestMeta{
			ID:     h.OperationID(),
			Method: h.Method(),
			Route:  h.Path(),
		},
	}
	m.operations.add(info)
	segments := h.PathSegments()

	operators := &strings.Builder{}
	_, _ = fmt.Fprintf(operators, "{{ ")
	for i, o := range h.Operators() {
		if i > 0 {
			_, _ = fmt.Fprintf(operators, " -> ")
		}
		_, _ = fmt.Fprintf(operators, "%s", o.String())
	}
	_, _ = fmt.Fprintf(operators, " }}")
	m.printer.Add(&meta{
		Method:    method,
		Path:      segments.PathString(),
		Summary:   h.Summary(),
		Operators: operators.String(),
	})

	ua := info.UA()
	injections := []contextx.Carrier{
		types.CarryOperationMeta(info),
		types.CarryOperationMetaProvider(m.operations),
	}

	r.HandleFunc(
		method+" "+path.Prefix(segments),
		func(rw http.ResponseWriter, req *http.Request) {
			injections = append(injections, path.CarryParamGetter(req))
			ctx := contextx.Compose(injections...)(req.Context())
			rw.Header().Set("Server", ua)
			h.ServeHTTP(rw, req.WithContext(ctx))
		},
	)
}

func (m *mux) build() http.Handler {
	m.printer = &printer{}
	defer m.printer.Flush()

	return m.group().handler(m)
}

func (m *mux) Handler() (http.Handler, error) {
	return m.build(), nil
}

type operations struct {
	spec *oas.OpenAPI
	meta map[string]types.OperationMeta
}

func (o *operations) add(meta *types.OperationMeta) {
	if o.meta == nil {
		o.meta = make(map[string]types.OperationMeta)
	}
	o.meta[meta.ID] = *meta
}

func (o *operations) GetOperation(id string) (types.OperationMeta, bool) {
	if o.meta == nil {
		return types.OperationMeta{}, false
	}
	info, ok := o.meta[id]
	return info, ok
}

func (o *operations) OpenAPI() *oas.OpenAPI {
	return o.spec
}
