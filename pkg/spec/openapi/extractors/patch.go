package extractors

import (
	"reflect"

	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/va"
	"github.com/xoctopus/httpx/pkg/spec/openapi/jsonschema"
	"github.com/xoctopus/httpx/pkg/validation/validators"
)

func PatchSchemaValidation(s jsonschema.Schema, opt validation.Option) (jsonschema.Schema, error) {
	va, err := validation.New(opt)
	if err != nil {
		return nil, err
	}
	if opt.Rule != nil {
		s = patchValidation(s, opt.Type, va)
		s.GetMetadata().AddExtension(jsonschema.XTagValidate, opt.Rule)
	}
	return s, nil
}

func patchValidation(s jsonschema.Schema, t reflect.Type, v validation.Validator) jsonschema.Schema {
	if u, ok := v.(interface{ Unwrap() validation.Validator }); ok {
		return patchValidation(s, t, u.Unwrap())
	}

	if x, ok := v.(*validators.UserDefined); ok {
		return &jsonschema.StringType{Type: "string", Format: x.Format()}
	}

	if x, ok := v.(va.EnumValidation); ok {
		return &jsonschema.EnumType{Enum: x.Enums()}
	}

	switch vt := v.(type) {
	case *validators.Int[int64]:
		return patchNumericMultiple[int64](patchNumericRange[int64](s, vt), vt)
	case *validators.Int[uint64]:
		return patchNumericMultiple[uint64](patchNumericRange[uint64](s, vt), vt)
	case *validators.Float:
		return patchNumericMultiple[float64](patchNumericRange[float64](s, vt), vt)
	case *validators.String:
		ss := patchListLength(jsonschema.String(), vt).(*jsonschema.StringType)
		ss.Format = vt.Format()
		if _va := vt.RegexpVa; _va != nil {
			ss.Pattern = _va.Pattern()
			if hint := _va.Hint(); len(hint) > 0 {
				ss.AddExtension(jsonschema.XPatternErrMsg, hint)
			}
		}
		return ss
	case *validators.Slice:
		switch x := s.(type) {
		case *jsonschema.ArrayType:
			x = patchListLength(x, vt).(*jsonschema.ArrayType)
			option := validation.Option{
				Type: t.Elem(),
				Rule: vt.ElemRule(),
			}
			x.Items = must.NoErrorV(PatchSchemaValidation(x.Items, option))
			return x
		}
	case *validators.Map:
		switch x := s.(type) {
		case *jsonschema.ObjectType:
			x = patchListLength(x, vt).(*jsonschema.ObjectType)
			optionk := validation.Option{Type: t.Key(), Rule: vt.KeyRule()}
			x.PropertyNames = must.NoErrorV(PatchSchemaValidation(x.PropertyNames, optionk))
			optionv := validation.Option{Type: t.Elem(), Rule: vt.ElemRule()}
			x.AdditionalProperties = must.NoErrorV(PatchSchemaValidation(x.AdditionalProperties, optionv))
			return x
		}
	}
	return s
}

func patchNumericRange[T va.Numeric](s jsonschema.Schema, va va.NumericRangeValidation[T]) jsonschema.Schema {
	if x, ok := s.(*jsonschema.NumberType); ok {
		if v := va.Min(); v != nil {
			if va.ExclusiveMin() {
				x.ExclusiveMinimum = new(float64(*v))
			} else {
				x.Minimum = new(float64(*v))
			}
		}
		if v := va.Max(); v != nil {
			if va.ExclusiveMax() {
				x.ExclusiveMaximum = new(float64(*v))
			} else {
				x.Maximum = new(float64(*v))
			}
		}
	}
	return s
}

func patchNumericMultiple[T va.Numeric](s jsonschema.Schema, va va.NumericMultipleValidation[T]) jsonschema.Schema {
	if x, ok := s.(*jsonschema.NumberType); ok {
		if v := va.MultipleOf(); v > 0 {
			x.MultipleOf = new(float64(v))
		}
	}
	return s
}

func patchListLength(s jsonschema.Schema, va va.LengthValidation) jsonschema.Schema {
	switch x := s.(type) {
	case *jsonschema.StringType:
		x.MinLength, x.MaxLength = va.LengthRange()
	case *jsonschema.ArrayType:
		x.MinItems, x.MaxItems = va.LengthRange()
	case *jsonschema.ObjectType:
		x.MinProperties, x.MaxProperties = va.LengthRange()
	}
	return s
}
