package validation

import (
	"encoding"
	"errors"
	"reflect"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/internal/validation/rule"
)

var (
	tTextMarshaler       = reflect.TypeFor[encoding.TextMarshaler]()
	tTextUnmarshaler     = reflect.TypeFor[encoding.TextUnmarshaler]()
	tJsonMarshaler       = reflect.TypeFor[json.Marshaler]()
	tJsonUnmarshaler     = reflect.TypeFor[json.Unmarshaler]()
	tJsonUnmarshalerFrom = reflect.TypeFor[json.UnmarshalerFrom]()
	tJsonTextValue       = reflect.TypeFor[jsontext.Value]()
	tBytes               = reflect.TypeFor[[]byte]()
)

func UnmarshalDecode(d *jsontext.Decoder, out any, o ...json.Options) error {
	var va Validator

	if x, ok := out.(scanner.WrappedField); ok {
		out = x.Unwrap()
		v, err := NewFromStructField(x.Field())
		if err != nil {
			return err
		}
		va = v
	}

	rv, ok := out.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(out)
	}

	if rv.Kind() != reflect.Pointer {
		return errors.New("must be pointer value for UnmarshalDecode")
	}

	if va == nil {
		if w, ok := out.(TagValidator); ok {
			if tag := w.ValidationTag(); tag != "" {
				r, err := rule.Compile(tag)
				if err != nil {
					return err
				}

				v, err := New(Option{Type: rv.Type(), Rule: r})
				if err != nil {
					return err
				}
				va = v
			}
		}
	}

	return (&pointer{Value: rv, Validator: va}).
		UnmarshalDecode(d, json.JoinOptions(o...))
}

type pointer struct {
	reflect.Value
	Validator
}

func (pv *pointer) UnmarshalDecode(d *jsontext.Decoder, o json.Options) error {
	if d.PeekKind() == jsontext.NULL {
		raw, err := d.ReadValue()
		if err != nil {
			return err
		}
		if va := pv.Validator; va != nil {
			if err = va.Validate(raw); err != nil {
				return WrapPosition(err, d.StackPointer())
			}
		}
		if pv.CanAddr() {
			pv.SetZero()
		}
		return nil
	}

	if pv.CanAddr() {
		rv := reflect.New(pv.Type().Elem())
		vv := &valuer{rv.Elem(), pv.Validator}
		if err := vv.UnmarshalDecode(d, o); err != nil {
			return err
		}
		pv.Set(rv)
		return nil
	}

	if pv.IsNil() {
		pv.Set(reflect.New(pv.Type().Elem()))
	}

	vv := &valuer{pv.Elem(), pv.Validator}
	if err := vv.UnmarshalDecode(d, o); err != nil {
		return err
	}
	return nil
}

type valuer struct {
	reflect.Value
	Validator
}

func (vv *valuer) UnmarshalDecode(d *jsontext.Decoder, o json.Options) error {
	if vv.Kind() == reflect.Pointer {
		return (&pointer{vv.Value, vv.Validator}).UnmarshalDecode(d, o)
	}

	typ := vv.Type()

	if reflect.PointerTo(typ).Implements(tJsonUnmarshalerFrom) {
		if err := json.UnmarshalDecode(d, vv.Addr().Interface(), o); err != nil {
			return WrapPosition(err, d.StackPointer())
		}
		return nil
	}

	if reflect.PointerTo(typ).Implements(tJsonUnmarshaler) {
		value, err := d.ReadValue()
		if err != nil {
			return err
		}
		u := vv.Addr().Interface().(json.Unmarshaler)
		if err = u.UnmarshalJSON(value); err != nil {
			return WrapPosition(err, d.StackPointer())
		}
		return nil
	}

	if reflect.PointerTo(typ).Implements(tTextUnmarshaler) ||
		typ == tJsonTextValue || typ == tBytes {
		return (&primitive{vv.Value, vv.Validator, true}).UnmarshalDecode(d, o)
	}

	switch typ.Kind() {
	case reflect.String:
		return (&primitive{vv.Value, vv.Validator, true}).UnmarshalDecode(d, o)
	case reflect.Bool:
		return (&primitive{vv.Value, vv.Validator, false}).UnmarshalDecode(d, o)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return (&primitive{vv.Value, vv.Validator, false}).UnmarshalDecode(d, o)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return (&primitive{vv.Value, vv.Validator, false}).UnmarshalDecode(d, o)
	case reflect.Float32, reflect.Float64:
		return (&primitive{vv.Value, vv.Validator, false}).UnmarshalDecode(d, o)
	case reflect.Struct:
		return (&compose{vv.Value, vv.Validator}).UnmarshalDecode(d, o)
	case reflect.Map:
		return (&pair{vv.Value, vv.Validator}).UnmarshalDecode(d, o)
	case reflect.Slice, reflect.Array:
		if reflect.SliceOf(vv.Type().Elem()).AssignableTo(tBytes) {
			return (&primitive{vv.Value, vv.Validator, true}).UnmarshalDecode(d, o)
		}
		return (&array{vv.Value, vv.Validator}).UnmarshalDecode(d, o)
	case reflect.Interface:
		return (&box{vv.Value, vv.Validator}).UnmarshalDecode(d, o)
	default:
		return &json.SemanticError{GoType: vv.Type()}
	}
}

type primitive struct {
	reflect.Value
	Validator
	text bool
}

func (pv *primitive) UnmarshalDecode(d *jsontext.Decoder, o json.Options) error {
	panic("todo")
}

type array struct {
	reflect.Value
	Validator
}

func (av *array) UnmarshalDecode(d *jsontext.Decoder, o json.Options) error {
	panic("todo")
}

type box struct {
	reflect.Value
	Validator
}

func (bv *box) UnmarshalDecode(d *jsontext.Decoder, o json.Options) error {
	panic("todo")
}

type compose struct {
	reflect.Value
	Validator
}

func (cv *compose) UnmarshalDecode(d *jsontext.Decoder, o json.Options) error {
	panic("todo")
}

type pair struct {
	reflect.Value
	Validator
}

func (kv *pair) UnmarshalDecode(d *jsontext.Decoder, o json.Options) error {
	panic("todo")
}
