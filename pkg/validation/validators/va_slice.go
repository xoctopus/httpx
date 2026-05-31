package validators

import (
	"reflect"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func init() {
	validation.Register(&_sliceP{})
}

type _sliceP struct{}

func (_sliceP) Name() string {
	return "slice"
}

func (_sliceP) Variants() []string {
	return []string{"slice", "array"}
}

func (c *_sliceP) New(r rule.Rule) (_ validation.Validator, err error) {
	v := &Slice{}

	if v.LengthVa, err = va.NewLengthVa(r); err != nil {
		return nil, err
	}

	params := r.Parameters()
	if len(params) == 0 {
		return v, nil
	}
	if len(params) != 1 {
		return nil, codex.Errorf(validation.ERROR__SLICE_PARAM, "expect one parameter for element")
	}
	if _, ok := params[0].(rule.Rule); !ok {
		return nil, codex.Errorf(validation.ERROR__SLICE_PARAM, "expect one rule for element")
	}
	v.elem = params[0].(rule.Rule)

	return v, nil
}

type Slice struct {
	elem rule.Rule

	*va.LengthVa
}

func (v *Slice) ElemRule() rule.Rule {
	return v.elem
}

func (v *Slice) Validate(value []byte) error {
	return nil
}

func (v *Slice) String() string {
	b := rule.NewBuilder("slice")

	b.AddParameters(v.elem)

	v.LengthVa.BuiltTo(b)

	return string(b.Bytes())
}

func (v *Slice) PostValidate(rv reflect.Value) error {
	length := uint(0)
	if rv.Kind() == reflect.Array || !rv.IsNil() {
		length = uint(rv.Len())
	}

	return v.LengthVa.Validate(length)
}
