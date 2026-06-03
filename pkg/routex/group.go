package routex

import (
	"bytes"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/xoctopus/httpx/internal/payload/path"
	"github.com/xoctopus/httpx/internal/route"
)

type group struct {
	part     path.Segments
	parent   *group
	children map[string]*group
	handlers []route.Handler
}

func (g *group) segments() []string {
	n := len(g.part)
	if n == 0 {
		return nil
	}
	named, ok := g.part[n-1].(path.NamedSegment)
	if !ok || !named.Multiple() {
		return nil
	}

	emitted := map[string]bool{}
	segments := make([]string, 0)
	for _, c := range g.children {
		if len(c.part) > 0 {
			seg := c.part[0].PathString()
			if emitted[seg] {
				continue
			}
			emitted[seg] = true
			segments = append(segments, seg)
		}
	}
	return segments
}

func (g *group) handler(m *mux) http.Handler {
	r := http.NewServeMux()

	if len(g.handlers) > 0 {
		for _, h := range g.handlers {
			m.add(r, h)
		}
		return r
	}

	keys := slices.Sorted(maps.Keys(g.children))
	for _, k := range keys {
		var (
			child    = g.children[k]
			hh       = child.handler(m)
			prefix   = child.prefix()
			segments = child.segments()
		)
		if len(segments) > 0 {
			r.HandleFunc(
				PathPrefix(prefix),
				func(rw http.ResponseWriter, req *http.Request) {
					values := path.Values{}
					remain, ok := prefix.MatchTo(values, req.URL.Path)
					if ok {
						parts := strings.Split(remain, "/")
						for i, p := range parts {
							if slices.Contains(segments, p) {
								multi := strings.Join(parts[0:i], "/")
								value := url.PathEscape(multi)
								r2 := req.Clone(req.Context())
								u := *r2.URL
								u.Path = strings.Replace(req.URL.Path, multi, value, 1)
								r2.RequestURI = u.RequestURI()
								r2.URL = &u
								hh.ServeHTTP(rw, r2)
								return
							}
						}
					}
					http.NotFound(rw, req)
				},
			)
			continue
		}
		r.Handle(PathPrefix(prefix), hh)
	}
	return r
}

func (g *group) String() string {
	b := bytes.NewBuffer(nil)

	d := g.depth()

	_, _ = fmt.Fprint(b, "\n")
	_, _ = fmt.Fprint(b, strings.Repeat("  ", d))
	_, _ = fmt.Fprint(b, g.part.PathString())

	for _, c := range g.children {
		_, _ = fmt.Fprint(b, c.String())
	}

	for _, h := range g.handlers {
		_, _ = fmt.Fprint(b, "\n")
		_, _ = fmt.Fprint(b, strings.Repeat("  ", d+1))
		_, _ = fmt.Fprint(b, h.Method())
		_, _ = fmt.Fprint(b, " ")
		_, _ = fmt.Fprint(b, h.PathSegments().PathString())
	}

	return b.String()
}

func (g *group) depth() int {
	if g.parent == nil {
		return 0
	}
	return g.parent.depth() + 1
}

func (g *group) child(part path.Segments) *group {
	if g.children == nil {
		g.children = map[string]*group{}
	}

	p := part.PathString()
	child, ok := g.children[p]
	if !ok {
		c := &group{part: part, parent: g}
		g.children[p] = c
		return c
	}

	return child
}

func (g *group) add(h route.Handler, chunk ...path.Segments) {
	if len(chunk) > 0 {
		g.child(chunk[0]).add(h, chunk[1:]...)
		return
	}
	g.handlers = append(g.handlers, h)
}

func (g *group) prefix() path.Segments {
	if p := g.parent; p != nil {
		return append(p.prefix(), g.part...)
	}
	return g.part
}
