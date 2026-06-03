package va

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/slicex"

	"github.com/xoctopus/httpx/internal/validation/rule"
)

func NewEnumVa[T Enum](r rule.Rule) (*EnumVa[T], error) {
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
	if len(values) == 0 {
		return nil, nil
	}
	if len(values) == 1 {
		if _, ok := any(*new(T)).(string); !ok {
			data := values[0].Bytes()
			if len(data) > 0 && data[0] == '%' {
				return nil, nil
			}
		}
	}
	es := &EnumVa[T]{mm: make(map[T]struct{})}
	for _, v := range values {
		vt, err := ParseEnumValue[T](string(v.Bytes()))
		if err != nil {
			return nil, codex.Wrapf(ERROR__INVALID_ENUM, err, "expect type %s got %s", reflect.TypeFor[T](), string(v.Bytes()))
		}
		es.mm[vt] = struct{}{}
	}
	es.vs = slices.Collect(maps.Keys(es.mm))
	slices.Sort(es.vs)
	return es, nil
}

type EnumVa[T comparable] struct {
	vs []T
	mm map[T]struct{}
}

func (es *EnumVa[T]) Enums() []any {
	if es == nil {
		return nil
	}
	return slicex.M(es.vs, func(from T) any { return from })
}

func (es *EnumVa[T]) Validate(v T) error {
	if es == nil {
		return nil
	}
	if _, ok := es.mm[v]; ok {
		return nil
	}
	return codex.Errorf(ERROR__OUT_OF_ENUMERATED_VALUES, "expect %v got %v", es.vs, v)
}

func (es *EnumVa[T]) BuiltTo(b rule.Builder) {
	if es != nil {
		values := make([]*rule.Literal, 0, len(es.vs))
		for _, v := range es.vs {
			values = append(values, rule.NewLiteral(fmt.Sprint(v)))
		}
		b.AppendValues(values)
	}
}
