package validators

import (
	"reflect"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func init() {
	validation.Register(&_mapP{})
}

type _mapP struct{}

func (_mapP) Name() string {
	return "map"
}

func (_mapP) Variants() []string {
	return []string{"map", "record"}
}

func (c *_mapP) New(r rule.Rule) (_ validation.Validator, err error) {
	v := &Map{}

	if v.LengthVa, err = va.NewLengthVa(r); err != nil {
		return nil, err
	}

	params := r.Parameters()
	if len(params) != 2 {
		return nil, codex.Errorf(validation.ERROR__MAP_PARAM, "expect 2 parameters for key and value")
	}
	_, ok1 := params[0].(rule.Rule)
	_, ok2 := params[2].(rule.Rule)
	if !ok1 || !ok2 {
		return nil, codex.Errorf(validation.ERROR__MAP_PARAM, "expect rule parameters for key and value")
	}

	v.key.Rule = params[0].(rule.Rule)
	v.ele.Rule = params[1].(rule.Rule)

	return wrap(v, r), nil
}

type Map struct {
	key validation.Option
	ele validation.Option

	*va.LengthVa
}

func (v *Map) Key() validation.Option {
	return v.key
}

func (v *Map) Elem() validation.Option {
	return v.ele
}

func (v *Map) Validate(_ []byte) error {
	return nil
}

func (v *Map) String() string {
	b := rule.NewBuilder("map")

	b.AddParameters(v.key.Rule)
	b.AddParameters(v.ele.Rule)

	v.LengthVa.BuiltTo(b)

	return string(b.Bytes())
}

func (v *Map) PostValidate(rv reflect.Value) error {
	length := uint(0)
	if !rv.IsNil() {
		length = uint(rv.Len())
	}

	return v.LengthVa.Validate(length)
}
