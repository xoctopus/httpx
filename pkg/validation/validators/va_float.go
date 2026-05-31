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
	validation.Register(&_floatP{})
}

type _floatP struct{}

func (_floatP) Name() string {
	return "float"
}

func (_floatP) Variants() []string {
	return []string{"double", "float32", "float64"}
}

func (p *_floatP) New(r rule.Rule) (_ validation.Validator, err error) {
	v := &Float{}

	switch r.Name() {
	case "float", "float32":
		v.BitSize = 32
	default:
		v.BitSize = 64
	}
	v.FloatScaleVa, err = va.NewFloatScaleVa(r)
	if err != nil {
		return nil, err
	}
	v.EnumVa, err = va.NewEnumVa[float64](r)
	if err != nil {
		return nil, err
	}
	v.RangeVa, err = va.NewRangeVa[float64](r)
	if err != nil {
		return nil, err
	}
	v.MultipleVa, err = va.NewMultipleVa[float64](r)
	if err != nil {
		return nil, err
	}
	return v, nil
}

type Float struct {
	BitSize uint

	*va.FloatScaleVa
	*va.EnumVa[float64]
	*va.RangeVa[float64]
	*va.MultipleVa[float64]
}

func (v *Float) String() string {
	r := rule.NewBuilder("float")

	v.FloatScaleVa.BuildTo(r)
	v.EnumVa.BuiltTo(r)
	v.RangeVa.BuiltTo(r)
	v.MultipleVa.BuiltTo(r)

	return string(r.Bytes())
}

func (v *Float) Validate(value []byte) error {
	if k := jsontext.Value(value).Kind(); k != jsontext.NUMBER {
		return codex.Errorf(validation.ERROR__INPUT_TYPE, "expect: float number got: %s", k)
	}

	text := string(value)
	_, p := va.ExtractFloatScales(text)

	val, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return codex.Wrapf(validation.ERROR__INPUT_VALUE, err, "expect float, got %s", text)
	}

	return errors.Join(
		v.FloatScaleVa.Validate(text),
		v.RangeVa.Validate(val),
		v.MultipleVa.Validate(val, p),
		v.EnumVa.Validate(val),
	)
}
