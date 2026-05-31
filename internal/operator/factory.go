package operator

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/xoctopus/x/misc/must"
)

func NewFactory(op Operator, last bool) *Factory {
	f := &Factory{
		Type:     TypeOfOperator(reflect.TypeOf(op)),
		IsLast:   last,
		Operator: op,
	}

	must.BeTrueF(f.Type.Kind() == reflect.Struct, "operator must be struct type")

	if _, ok := op.(NoOutput); ok {
		f.NoOutput = true
	}

	if x, ok := op.(HasParameters); ok {
		f.Params = x.OperatorParameters()
	}

	if !f.IsLast {
		f.ContextKey = f.Type.String()
		if x, ok := op.(HasContextKey); ok {
			f.ContextKey = x.ContextKey()
		}
	}

	return f
}

type Factory struct {
	Type       reflect.Type
	ContextKey any
	NoOutput   bool
	Params     url.Values
	IsLast     bool
	Operator   Operator
}

func (f *Factory) String() string {
	s := ""
	if st, ok := f.Operator.(fmt.Stringer); ok {
		s = st.String()
	} else {
		s = f.Type.String()
	}

	if f.Params != nil {
		return s + "?" + f.Params.Encode()
	}

	return s
}

func (f *Factory) New() (op Operator) {
	if x, ok := f.Operator.(Newer); ok {
		op = x.New()
	} else {
		op = reflect.New(f.Type).Interface().(Operator)
	}

	if x, ok := op.(Initializer); ok {
		x.InitFromOperator(f.Operator)
	}

	if x, ok := op.(DefaultSetter); ok {
		x.SetDefault()
	}

	return op
}
