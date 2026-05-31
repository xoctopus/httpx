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
	if len(params) == 0 {
		return v, nil
	}

	if len(params) != 2 {
		return nil, codex.Errorf(validation.ERROR__MAP_PARAM, "expect 2 parameters for key and value")
	}
	kr, vr := false, false
	v.keyRule, kr = params[0].(rule.Rule)
	v.eleRule, vr = params[1].(rule.Rule)
	if !kr || !vr {
		return nil, codex.Errorf(validation.ERROR__MAP_PARAM, "expect rule parameters for key and value")
	}

	v.keyRule = params[0].(rule.Rule)
	v.eleRule = params[1].(rule.Rule)

	return v, nil
}

type Map struct {
	keyRule rule.Rule
	eleRule rule.Rule

	*va.LengthVa
}

func (v *Map) KeyRule() rule.Rule {
	return v.keyRule
}

func (v *Map) SetKeyRule(r rule.Rule) {
	v.keyRule = r
}

func (v *Map) ElemRule() rule.Rule {
	return v.eleRule
}

func (v *Map) SetElemRule(r rule.Rule) {
	v.eleRule = r
}

func (v *Map) Validate(_ []byte) error {
	return nil
}

func (v *Map) String() string {
	b := rule.NewBuilder("map")

	b.AddParameters(v.keyRule)
	b.AddParameters(v.eleRule)

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
