package operator

import (
	"context"
	"fmt"
	"path"
	"reflect"
)

// PathDescriber reports a path segment appended when building route metadata.
type PathDescriber interface {
	Path() string
}

// BasePathDescriber reports a base path prefix for a route group.
type BasePathDescriber interface {
	BasePath() string
}

// Operator is one step in a route pipeline. Output returns the value passed to
// the next step (via context) or, on the last step, the HTTP response body.
type Operator interface {
	Output(context.Context) (any, error)
}

// NoOutput marks a step that does not produce a response body by itself.
// [Factory] sets NoOutput when the registered prototype implements this interface.
type NoOutput interface {
	Operator
	NoOutput()
}

// Newer replaces the default reflect.New instance factory in [Factory.New].
type Newer interface {
	Operator
	New() Operator
}

// Initializer runs after request unmarshaling and before Output on each request.
type Initializer interface {
	Init(context.Context) error
}

// InitializerFrom copies or derives state from the registration prototype into
// the per-request instance inside [Factory.New].
type InitializerFrom interface {
	Operator
	InitFromOperator(Operator)
}

// DefaultSetter runs after instance creation in [Factory.New].
type DefaultSetter interface {
	SetDefault()
}

// HasParameters attaches static query parameters to [Factory] (for example client defaults).
type HasParameters interface {
	OperatorParameters() map[string][]string
}

// HasContextKey overrides the context key used when a non-final step's Output
// is stored for downstream operators. The default is the operator struct type name.
type HasContextKey interface {
	Operator
	ContextKey() any
}

// HasMiddlewares expands one router entry into multiple operators when building
// a [route.Router].
type HasMiddlewares interface {
	Middlewares() []Operator
}

// EmptyOperator is a no-op Output used as the embedded base for meta operators.
type EmptyOperator struct{}

func (EmptyOperator) NoOutput() {}

func (EmptyOperator) Output(context.Context) (any, error) { return nil, nil }

// BasePathOperator returns a meta operator that only contributes BasePath to route metadata.
func BasePathOperator(base string) Operator {
	return &MetaOperator{base: path.Clean(base)}
}

// GroupOperator returns a meta operator that only contributes Path to route metadata.
func GroupOperator(group string) Operator {
	return &MetaOperator{path: path.Clean(group)}
}

// MetaOperator contributes path metadata and performs no business Output.
type MetaOperator struct {
	path string
	base string
	EmptyOperator
}

func (g *MetaOperator) Path() string {
	return g.path
}

func (g *MetaOperator) BasePath() string {
	return g.base
}

func (g *MetaOperator) String() string {
	if len(g.base) > 0 {
		return fmt.Sprintf("base(%s)", g.base)
	}
	return fmt.Sprintf("group(%s)", g.path)
}

// TypeOfOperator returns the struct type for an operator, unwrapping pointers.
func TypeOfOperator(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		return TypeOfOperator(t.Elem())
	}
	return t
}
