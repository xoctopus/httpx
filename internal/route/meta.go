package route

import (
	"reflect"
	"strings"

	"github.com/xoctopus/x/docx"

	"github.com/xoctopus/httpx/internal/method"
	"github.com/xoctopus/httpx/internal/operator"
	"github.com/xoctopus/httpx/internal/payload/path"
)

type Meta struct {
	OperationID string
	Method      string
	Path        string
	BasePath    string
	Summary     string
	Description string
	Deprecated  bool
}

func NewMeta(f *operator.Factory) *Meta {
	m := &Meta{}

	op := f.Operator

	m.OperationID = f.Type.Name()

	if x, ok := op.(method.Describer); ok {
		m.Method = x.Method()
	}

	if x, ok := op.(docx.Doc); ok {
		if docs, ok := x.DocOf(); ok && len(docs) > 0 {
			m.Summary = docs[0]
			m.Description = strings.Join(docs[1:], "\n")
		}
	}

	if f.Type.Kind() == reflect.Struct {
		m.Path, m.Summary = path.ResolveFromTag(f.Type)
	}

	if x, ok := op.(operator.BasePathDescriber); ok {
		m.BasePath = x.BasePath()
	}

	if x, ok := op.(operator.PathDescriber); ok {
		m.Path = x.Path()
	}

	return m
}
