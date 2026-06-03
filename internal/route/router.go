package route

import (
	"bytes"
	"iter"
	"slices"
	"sort"
	"strings"

	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/httpx/internal/operator"
)

type Router interface {
	Register(r Router)
	With(routers ...Router) Router
	Routes() Routes
}

func NewRouter(operators ...operator.Operator) Router {
	return &router{
		ops: func(yield func(operator.Operator) bool) {
			for i := range operators {
				op := operators[i]

				if x, ok := op.(operator.HasMiddlewares); ok {
					for _, o := range x.Middlewares() {
						if !yield(o) {
							return
						}
					}
				}

				if !yield(op) {
					return
				}
			}
		},
	}
}

type router struct {
	prev *router
	next map[*router]bool
	ops  iter.Seq[operator.Operator]
}

func (rt router) With(routers ...Router) Router {
	next := &rt
	for i := range routers {
		next.Register(routers[i])
	}
	return next
}

func (rt *router) Register(r Router) {
	if rt.next == nil {
		rt.next = map[*router]bool{}
	}
	must.BeTrueF(
		r.(*router).prev == nil,
		"router is already registered. prev: %v, self: %v", rt.prev, r,
	)
	r.(*router).prev = rt
	rt.next[r.(*router)] = true
}

func (rt *router) route() *route {
	prev, ops := rt.prev, slices.Collect(rt.ops)

	for prev != nil {
		ops = append(slices.Collect(prev.ops), ops...)
		prev = prev.prev
	}

	return &route{
		ops:  ops,
		last: len(rt.next) == 0,
	}
}

func (rt *router) Routes() (results Routes) {
	maybeAppendRoute := func(router *router) {
		r := router.route()

		if r.last && len(r.ops) > 0 {
			results = append(results, r)
		}

		if len(router.next) > 0 {
			results = append(results, router.Routes()...)
		}
	}

	if len(rt.next) == 0 {
		maybeAppendRoute(rt)
		return results
	}

	for next := range rt.next {
		maybeAppendRoute(next)
	}

	return results
}

type Route interface {
	String() string
	Range(forEach func(op *operator.Factory, index int) error) error
}

type route struct {
	ops  []operator.Operator
	last bool
}

func (r *route) Range(each func(f *operator.Factory, i int) error) error {
	for i, op := range r.ops {
		if err := each(operator.NewFactory(op, i == len(r.ops)-1), i); err != nil {
			return err
		}
	}
	return nil
}

func (r *route) String() string {
	buf := &bytes.Buffer{}
	_ = r.Range(func(f *operator.Factory, i int) error {
		if i > 0 {
			buf.WriteString(" => ")
		}
		buf.WriteString(f.String())
		return nil
	})
	return buf.String()
}

type Routes []Route

func (routes Routes) String() string {
	keys := make([]string, len(routes))
	for i, r := range routes {
		keys[i] = r.String()
	}
	sort.Strings(keys)
	return strings.Join(keys, "\n")
}
