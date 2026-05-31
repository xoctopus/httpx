package va

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation/rule"
)

func NewRangeVa[T Numeric](r rule.Rule) (*RangeVa[T], error) {
	if r.LengthMode() {
		return nil, codex.Errorf(ERROR__INVALID_VALUE_RANGE, "expect range but got length mode")
	}
	minimum, exl := r.Min()
	maximum, exr := r.Max()
	if rule.IsNil(minimum) && rule.IsNil(maximum) {
		return nil, nil
	}

	vr := &RangeVa[T]{exMin: exl, exMax: exr}
	if !rule.IsNil(minimum) {
		v, err := ParseEnumValue[T](minimum.String())
		if err != nil {
			return nil, codex.Wrapf(ERROR__INVALID_VALUE_RANGE, err, "expect type %s, got %s", reflect.TypeFor[T](), minimum.String())
		}
		vr.min = new(v)
	}

	if !rule.IsNil(maximum) {
		v, err := ParseEnumValue[T](maximum.String())
		if err != nil {
			return nil, codex.Wrapf(ERROR__INVALID_VALUE_RANGE, err, "expect type %s, got %s", reflect.TypeFor[T](), maximum.String())
		}
		vr.max = new(v)
	}

	if vr.min != nil && vr.max != nil && *vr.max <= *vr.min {
		return nil, codex.Errorf(ERROR__INVALID_VALUE_RANGE, "min>max %d>%d", *vr.min, *vr.max)
	}

	return vr, nil
}

type RangeVa[T Numeric] struct {
	min   *T
	exMin bool
	max   *T
	exMax bool
}

func (vr *RangeVa[T]) BuiltTo(b rule.Builder) {
	if vr != nil {
		_min, _max := "", ""
		if vr.min != nil {
			_min = fmt.Sprint(*vr.min)
		}
		if vr.max != nil {
			_max = fmt.Sprint(*vr.max)
		}
		b.SetMin(rule.NewLiteral(_min), vr.exMin)
		b.SetMax(rule.NewLiteral(_max), vr.exMax)
	}
}

func (vr *RangeVa[T]) Validate(v T) error {
	if vr == nil {
		return nil
	}
	if vr.min != nil && (v < *vr.min || vr.exMin && v <= *vr.min) ||
		vr.max != nil && (v > *vr.max || vr.exMax && v >= *vr.max) {
		return codex.Errorf(ERROR__OUT_OF_VALUE_RANGE, "expect %s got %v", vr, v)
	}
	return nil
}

func (vr *RangeVa[T]) String() string {
	if vr == nil {
		return ""
	}
	b := bytes.NewBuffer(nil)
	if vr.exMin {
		b.WriteByte('(')
	} else {
		b.WriteByte('[')
	}
	if vr.min != nil {
		b.WriteString(fmt.Sprint(*vr.min))
	}
	b.WriteByte(',')
	if vr.max != nil {
		b.WriteString(fmt.Sprint(*vr.max))
	}
	if vr.exMax {
		b.WriteByte(')')
	} else {
		b.WriteByte(']')
	}
	return b.String()
}
