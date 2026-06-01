package path

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/httpx/internal/types"
)

func SplitPath(p string) []string {
	p = CleanPath(p)
	if p[0] == '/' {
		p = p[1:]
	}
	if p == "" {
		return make([]string, 0)
	}
	return strings.Split(p, "/")
}

func CleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	if p[len(p)-1] == '/' && np != "/" {
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

func Normalize(p string) string {
	parts := SplitPath(path.Clean(p))
	processed := make([]string, len(parts))

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			processed[i] = fmt.Sprintf("{%s}", part[1:])
			continue
		}
		if strings.HasPrefix(part, "*") {
			processed[i] = fmt.Sprintf("{%s...}", part[1:])
			continue
		}

		processed[i] = part
	}

	return "/" + strings.Join(processed, "/")
}

func ResolveFromTag(t reflect.Type) (string, string) {
	for f := range t.Fields() {
		if f.Anonymous {
			if f.Type.PkgPath() == types.ExposedRoot && strings.HasPrefix(f.Name, "Method") {
				p, summary := "", ""
				if x, ok := f.Tag.Lookup("path"); ok {
					p = Normalize(x)
				}
				if x, ok := f.Tag.Lookup("summary"); ok {
					summary = x
				}
				return p, summary
			}
			// deep walk
			if f.Type.Kind() == reflect.Struct {
				return ResolveFromTag(f.Type)
			}
		}
	}
	return "", ""
}

type (
	tCtxValueGetter struct{}

	ValueGetter interface {
		PathValue(k string) string
	}
)

var WithParamGetter = contextx.With[tCtxValueGetter, ValueGetter]

func ParamGetterFrom(ctx context.Context) ValueGetter {
	if x, ok := contextx.From[tCtxValueGetter, ValueGetter](ctx); ok && x != nil {
		return x
	}
	return Values{}
}

type Values map[string]string

func (vs Values) PathValue(k string) string {
	return vs[k]
}

func (vs Values) SetPathValue(key, value string) {
	vs[key] = value
}

type Encoder interface {
	Encode(Values) string
}

type ValuesDescriber interface {
	PathValues(string) (Values, error)
}

type ValuesModifier interface {
	SetPathValue(key, val string)
}
