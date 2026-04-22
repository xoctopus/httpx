package validators

import (
	"errors"
	"strconv"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func init() {
	validation.Register(&_integerP{})
}

type _integerP struct{}

func (_integerP) Name() string {
	return "int"
}

func (_integerP) Variants() []string {
	variants := make([]string, 0, 130)
	variants = append(variants, "int", "uint")
	for i := 1; i <= 64; i++ {
		bits := strconv.Itoa(i)
		variants = append(variants, "int"+bits, "uint"+bits)
	}
	return variants
}

func (p *_integerP) New(r rule.Rule) (validation.Validator, error) {
	scale, err := va.NewIntScaleVa(r)
	if err != nil {
		return nil, err
	}

	if scale.Unsigned() {
		v := &Int[uint64]{
			IntScaleVa: scale,
		}
		if v.RangeVa, err = va.NewRangeVa[uint64](r); err != nil {
			return nil, err
		}
		if v.EnumVa, err = va.NewEnumVa[uint64](r); err != nil {
			return nil, err
		}
		if v.MultipleVa, err = va.NewMultipleVa[uint64](r); err != nil {
			return nil, err
		}
		return v, nil
	}
	v := &Int[int64]{
		IntScaleVa: scale,
	}
	if v.RangeVa, err = va.NewRangeVa[int64](r); err != nil {
		return nil, err
	}
	if v.EnumVa, err = va.NewEnumVa[int64](r); err != nil {
		return nil, err
	}
	if v.MultipleVa, err = va.NewMultipleVa[int64](r); err != nil {
		return nil, err
	}
	return wrap(v, r), nil
}

type Int[T int64 | uint64] struct {
	*va.IntScaleVa
	*va.EnumVa[T]
	*va.RangeVa[T]
	*va.MultipleVa[T]
}

func (v *Int[T]) String() string {
	name := "int"
	if v.Unsigned() {
		name = "uint"
	}
	rb := rule.NewBuilder(name)

	v.IntScaleVa.BuiltTo(rb)
	v.EnumVa.BuiltTo(rb)
	v.MultipleVa.BuiltTo(rb)
	v.RangeVa.BuiltTo(rb)

	return string(rb.Bytes())
}

func (v *Int[T]) Validate(value []byte) error {
	k := jsontext.Value(value).Kind()
	if k != jsontext.NUMBER {
		return codex.Errorf(validation.ERROR__INPUT_TYPE, "expect: number got: %s", k)
	}

	x := *new(T)
	switch any(x).(type) {
	case uint64:
		u, err := strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return codex.Wrap(validation.ERROR__INPUT_VALUE, err)
		}
		x = T(u)
	default:
		u, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return codex.Wrap(validation.ERROR__INPUT_VALUE, err)
		}
		x = T(u)
	}
	return errors.Join(
		v.RangeVa.Validate(x),
		v.IntScaleVa.Validate(x),
		v.EnumVa.Validate(x),
		v.MultipleVa.Validate(x, 0),
	)
}
