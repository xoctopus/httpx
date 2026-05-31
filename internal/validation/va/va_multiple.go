package va

import (
	"fmt"
	"math"
	"reflect"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/validation/rule"
)

func NewMultipleVa[T Numeric](r rule.Rule) (*MultipleVa[T], error) {
	matrix := r.ValueMatrix()
	if len(matrix) == 0 {
		return nil, nil
	}
	values := make([]*rule.Literal, 0, len(matrix[0]))
	for _, v := range matrix[0] {
		if !rule.IsNil(v) {
			values = append(values, v)
		}
	}
	if len(values) != 1 {
		return nil, nil
	}
	data := values[0].Bytes()
	if len(data) == 0 || data[0] != '%' {
		return nil, nil
	}
	div, err := ParseEnumValue[T](string(data[1:]))
	if err != nil {
		return nil, codex.Wrapf(ERROR__INVALID_MULTIPLE, err, "expect type %s got %s", reflect.TypeFor[T](), string(data[1:]))
	}
	if div == T(0) {
		return nil, codex.Errorf(ERROR__INVALID_MULTIPLE, "for multiple value should not equal 0")
	}
	return &MultipleVa[T]{div: div}, nil
}

type MultipleVa[T Numeric] struct {
	div T
}

func (m *MultipleVa[T]) BuiltTo(b rule.Builder) {
	if m != nil && m.div != T(0) {
		b.AppendValues([]*rule.Literal{rule.NewLiteral("%" + fmt.Sprint(m.div))})
	}
}

func (m *MultipleVa[T]) Validate(v T, precision uint) error {
	if m == nil || m.div == T(0) {
		return nil
	}
	multiple := false
	switch k := reflect.TypeFor[T]().Kind(); {
	case k >= reflect.Int && k <= reflect.Int64:
		vv, div := reflect.ValueOf(v).Int(), reflect.ValueOf(m.div).Int()
		multiple = vv%div == 0
	case k >= reflect.Uint && k <= reflect.Uint64:
		vv, div := reflect.ValueOf(v).Uint(), reflect.ValueOf(m.div).Uint()
		multiple = vv%div == 0
	default:
		vv, div := reflect.ValueOf(v).Float(), reflect.ValueOf(m.div).Float()
		remains := math.Abs(math.Mod(vv, div))
		fraction := math.Pow10(0 - int(precision)) // precision = 3 => frac = 0.001
		multiple = remains < fraction || math.Abs(remains-math.Abs(div)) < fraction
	}
	if multiple {
		return nil
	}
	return codex.Errorf(ERROR__NOT_MATCH_MULTIPLE, "expect multiple of %v got %v", m.div, v)
}
