package operator

import (
	"context"
	"fmt"
	"path"
	"reflect"
)

type PathDescriber interface {
	Path() string
}

type BasePathDescriber interface {
	BasePath() string
}

type Operator interface {
	Output(context.Context) (any, error)
}

type NoOutput interface {
	Operator
	NoOutput()
}

type Newer interface {
	Operator
	New() Operator
}

type Initializer interface {
	Init(context.Context) error
}

type InitializerFrom interface {
	Operator
	InitFromOperator(Operator)
}

type DefaultSetter interface {
	SetDefault()
}

type HasParameters interface {
	OperatorParameters() map[string][]string
}

type HasContextKey interface {
	Operator
	ContextKey() any
}

type HasMiddlewares interface {
	Middlewares() []Operator
}

type EmptyOperator struct{}

func (EmptyOperator) NoOutput() {}

func (EmptyOperator) Output(context.Context) (any, error) { return nil, nil }

func BasePathOperator(base string) Operator {
	return &MetaOperator{base: path.Clean(base)}
}

func GroupOperator(group string) Operator {
	return &MetaOperator{path: path.Clean(group)}
}

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

func TypeOfOperator(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		return TypeOfOperator(t.Elem())
	}
	return t
}
