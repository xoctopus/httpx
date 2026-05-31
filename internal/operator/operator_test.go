package operator_test

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	. "github.com/xoctopus/httpx/internal/operator"
)

func ExampleMetaOperator() {
	op := BasePathOperator("/base")
	fmt.Println(op.(BasePathDescriber).BasePath())
	fmt.Println(op)

	op = GroupOperator("/base/group")
	fmt.Println(op.(PathDescriber).Path())
	fmt.Println(op)

	// Output:
	// /base
	// base(/base)
	// /base/group
	// group(/base/group)
}

type OpX struct{ V int }

func (x *OpX) Output(ctx context.Context) (any, error) {
	return "x", nil
}

type OpY struct{}

func (y *OpY) Output(ctx context.Context) (any, error) {
	return "y", nil
}

type OpZ struct{}

func (z *OpZ) Output(ctx context.Context) (any, error) {
	return "z", nil
}

type OpHasParameters struct{ OpX }

func (x *OpHasParameters) OperatorParameters() map[string][]string {
	return map[string][]string{"input": []string{"100"}}
}

type OpHasContextKey struct{ OpX }

func (x *OpX) ContextKey() any {
	type _opk struct{}
	return _opk{}
}

type OpNoOutput struct{ OpX }

func (*OpNoOutput) NoOutput() {}

type OpCanBeStringified struct{ OpX }

func (*OpCanBeStringified) String() string {
	return "OpCanBeStringified__"
}

type OpCanBeInitialized struct{ OpX }

func (op *OpCanBeInitialized) InitFromOperator(from Operator) {
	if x, ok := from.(*OpCanBeInitialized); ok {
		op.OpX = x.OpX
		op.V = 300
	}
}

type OpHasMiddlewares struct{ OpX }

func (*OpHasMiddlewares) Middlewares() []Operator {
	return []Operator{&OpY{}, &OpZ{}}
}

type OpCanSetDefaults struct{ OpX }

func (x *OpCanSetDefaults) SetDefault() { x.V = 100 }

type OpProvider struct{ OpX }

func (x *OpProvider) New() Operator {
	o := &OpProvider{}
	o.V = 200
	return o
}

func TestNewFactory(t *testing.T) {
	f := NewFactory(&OpX{}, false)
	Expect(t, f.Type, Equal(reflect.TypeFor[OpX]()))
	Expect(t, f.ContextKey, NotBeNil[any]())
	Expect(t, f.String(), HaveSuffix("OpX"))

	f = NewFactory(&OpX{}, true)
	Expect(t, f.ContextKey, BeNil[any]())

	f = NewFactory(&OpCanBeStringified{}, false)
	Expect(t, f.String(), Equal("OpCanBeStringified__"))

	f = NewFactory(&OpHasParameters{}, false)
	Expect(t, f.Params, Equal(url.Values{"input": {"100"}}))
	Expect(t, f.String(), HaveSuffix("OpHasParameters?input=100"))

	f = NewFactory(&OpCanSetDefaults{}, false)
	x := f.New()
	o1, ok := x.(*OpCanSetDefaults)
	Expect(t, ok, BeTrue())
	Expect(t, o1.V, Equal(100))

	f = NewFactory(&OpNoOutput{}, false)
	Expect(t, f.NoOutput, BeTrue())

	f = NewFactory(&OpProvider{}, false)
	x = f.New()
	o2, ok := x.(*OpProvider)
	Expect(t, ok, BeTrue())
	Expect(t, o2.V, Equal(200))

	f = NewFactory(&OpCanBeInitialized{}, false)
	x = f.New()
	o3, ok := x.(*OpCanBeInitialized)
	Expect(t, ok, BeTrue())
	Expect(t, o3.V, Equal(300))
}
