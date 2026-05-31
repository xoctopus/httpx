package validation

import (
	"bytes"
	"encoding"
	"reflect"

	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/misc/must"

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
	tPlaceholder         = reflect.TypeFor[struct{}]()
)

type UnmarshalDecoder interface {
	UnmarshalDecode(d *jsontext.Decoder, o ...json.Options) error
}

func UnmarshalDecode(d *jsontext.Decoder, v any, o ...json.Options) error {
	var va Validator

	if x, ok := v.(scanner.WrappedField); ok {
		v, va = x.Unwrap(), must.NoErrorV(NewFromStructField(x.Field()))
	}

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	must.BeTrueWrap(rv.Kind() == reflect.Pointer, codex.New(ERROR__DEC_INVALID_INPUT))

	if va == nil {
		if x, ok := v.(TagValidator); ok {
			if tag := x.ValidationTag(); tag != "" {
				r, err := rule.Compile(tag)
				if err != nil {
					return err
				}
				_va, err := New(Option{Type: rv.Type(), Rule: r})
				if err != nil {
					return err
				}
				va = _va
			}
		}
	}

	return Pointer(rv, va).UnmarshalDecode(d, json.JoinOptions(o...))
}

func Pointer(rv reflect.Value, va Validator) UnmarshalDecoder {
	return &pointer{Value: rv, Validator: va}
}

type pointer struct {
	reflect.Value
	Validator
}

func (pv *pointer) UnmarshalDecode(d *jsontext.Decoder, o ...json.Options) error {
	if d.PeekKind() == jsontext.NULL {
		raw, err := d.ReadValue()
		if err != nil {
			return err
		}
		pos := d.StackPointer()
		if va := pv.Validator; va != nil {
			if err = va.Validate(raw); err != nil {
				return WrapPositionError(err, pos)
			}
		}
		if pv.CanSet() {
			pv.SetZero()
		}
		return nil
	}

	if pv.CanSet() {
		rv := reflect.New(pv.Type().Elem())
		if err := Value(rv.Elem(), pv.Validator).UnmarshalDecode(d, o...); err != nil {
			return err
		}
		pv.Set(rv)
		return nil
	}

	// if pv.IsNil() {
	// 	pv.Set(reflect.New(pv.Type().Elem()))
	// }

	return Value(pv.Elem(), pv.Validator).UnmarshalDecode(d, o...)
}

func Value(rv reflect.Value, va Validator) UnmarshalDecoder {
	return &value{Value: rv, Validator: va}
}

type value struct {
	reflect.Value
	Validator
}

func (vv *value) UnmarshalDecode(d *jsontext.Decoder, o ...json.Options) error {
	if vv.Kind() == reflect.Pointer {
		return Pointer(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	}

	typ := vv.Type()
	pos := d.StackPointer()

	if reflect.PointerTo(typ).Implements(tJsonUnmarshalerFrom) {
		if err := json.UnmarshalDecode(d, vv.Addr().Interface(), o...); err != nil {
			return WrapPositionError(err, pos)
		}
		return nil
	}

	if reflect.PointerTo(typ).Implements(tJsonUnmarshaler) {
		val, err := d.ReadValue()
		if err != nil {
			return err
		}
		raw := string(val)
		u := vv.Addr().Interface().(json.Unmarshaler)
		if err = u.UnmarshalJSON([]byte(raw)); err != nil {
			return WrapPositionError(err, pos)
		}
		return nil
	}

	if reflect.PointerTo(typ).Implements(tTextUnmarshaler) ||
		typ == tJsonTextValue || typ == tBytes {
		return Text(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	}

	switch typ.Kind() {
	case reflect.String:
		return Text(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	case reflect.Bool:
		return Basic(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Basic(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Basic(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	case reflect.Float32, reflect.Float64:
		return Basic(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	case reflect.Struct:
		return Struct(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	case reflect.Map:
		return Map(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	case reflect.Slice, reflect.Array:
		if reflect.SliceOf(vv.Type().Elem()).AssignableTo(tBytes) {
			return Text(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
		}
		return Array(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	case reflect.Interface:
		return Any(vv.Value, vv.Validator).UnmarshalDecode(d, o...)
	default:
		return &json.SemanticError{GoType: vv.Type()}
	}
}

func Basic(rv reflect.Value, va Validator) UnmarshalDecoder {
	return &primitive{Value: rv, Validator: va, text: false}
}

func Text(rv reflect.Value, va Validator) UnmarshalDecoder {
	return &primitive{Value: rv, Validator: va, text: true}
}

type primitive struct {
	reflect.Value
	Validator
	text bool
}

func (pv *primitive) UnmarshalDecode(d *jsontext.Decoder, o ...json.Options) error {
	var (
		tok = d.PeekKind()
		val jsontext.Value
		raw string
		pos jsontext.Pointer
		err error
	)
	switch tok {
	case jsontext.ARRAY:
		pos = d.StackPointer()
		if _, err = d.ReadToken(); err != nil {
			return err
		}
		// first array element for unmarshalling
		if val, err = d.ReadValue(); err != nil {
			return err
		}
		raw = val.String()

		// skip others
		for d.PeekKind() != jsontext.ARRAY_E {
			if _, err = d.ReadValue(); err != nil {
				return err
			}
		}
		if _, err = d.ReadToken(); err != nil {
			return err
		}
	default:
		pos = d.StackPointer()
		if val, err = d.ReadValue(); err != nil {
			return err
		}
		raw = val.String()
	}

	val = jsontext.Value(raw)
	if val.Kind() == jsontext.STRING {
		if !pv.text {
			if val, err = jsontext.AppendUnquote(nil, val); err != nil {
				return err
			}
		}
	}
	if va := pv.Validator; va != nil {
		if err = va.Validate(val); err != nil {
			return WrapPositionError(err, pos)
		}
	}
	if err = json.Unmarshal(val, pv.Value.Addr().Interface(), o...); err != nil {
		return WrapPositionError(err, pos)
	}
	return nil
}

func Array(rv reflect.Value, va Validator) UnmarshalDecoder {
	return &array{Value: rv, Validator: va}
}

type array struct {
	reflect.Value
	Validator
}

func (av *array) SetLen(n int) {
	if av.Kind() == reflect.Slice {
		av.Value.SetLen(n)
	}
}

func (av *array) UnmarshalDecode(d *jsontext.Decoder, o ...json.Options) error {
	tok, err := d.ReadToken()
	if err != nil {
		return err
	}
	k := tok.Kind()
	switch k {
	case jsontext.NULL:
		pos := d.StackPointer()
		if va := av.Validator; va != nil {
			if err = va.Validate([]byte("null")); err != nil {
				return WrapPositionError(err, pos)
			}
		}
		av.SetZero()
		return nil
	case jsontext.ARRAY:
		var eva Validator // element validator

		if av.Cap() > 0 {
			av.SetLen(av.Cap())
		}

		if x, ok := av.Validator.(WithElemRule); ok {
			option := Option{
				Type: av.Type().Elem(),
				Rule: x.ElemRule(),
			}
			if eva, err = New(option); err != nil {
				return err
			}
		}

		i, zero := 0, true
		for d.PeekKind() != jsontext.ARRAY_E {
			if i == av.Cap() {
				av.Grow(1)
				av.SetLen(av.Cap())
				zero = false
			}
			elem := &value{Value: av.Index(i), Validator: eva}
			i++
			if zero {
				elem.SetZero()
			}
			if err = elem.UnmarshalDecode(d, o...); err != nil {
				// if !IsValidationError(err) {
				// 	av.SetLen(i)
				// 	return &json.SemanticError{Err: err, GoType: av.Type()}
				// }
				return err
			}
		}
		if _, err = d.ReadToken(); err != nil {
			return err
		}
		if post, ok := av.Validator.(PostValidator); ok {
			if err = post.PostValidate(av.Value); err != nil {
				return WrapPositionError(err, d.StackPointer())
			}
		}
		if i == 0 {
			if av.Kind() == reflect.Slice {
				av.Set(reflect.MakeSlice(av.Type(), 0, 0))
			}
		} else {
			av.SetLen(i)
		}
		return nil
	default:
		return &json.SemanticError{JSONKind: k, GoType: av.Type()}
	}
}

func Any(rv reflect.Value, va Validator) UnmarshalDecoder {
	return &_any{Value: rv, Validator: va}
}

type _any struct {
	reflect.Value
	Validator
}

func (bv *_any) UnmarshalDecode(d *jsontext.Decoder, o ...json.Options) error {
	if d.PeekKind() == jsontext.NULL {
		// TODO should need to check optional
		if _, err := d.ReadToken(); err != nil {
			return err
		}
		bv.SetZero()
		return nil
	}

	var x any
	if err := json.UnmarshalDecode(d, &x, o...); err != nil {
		return err
	}
	bv.Set(reflect.ValueOf(x))
	return nil
}

func Struct(rv reflect.Value, va Validator) UnmarshalDecoder {
	return &_struct{Value: rv, Validator: va}
}

type _struct struct {
	reflect.Value
	Validator
}

func (cv *_struct) UnmarshalDecode(d *jsontext.Decoder, o ...json.Options) error {
	var k jsontext.Kind

	if tok, err := d.ReadToken(); err != nil {
		return err
	} else {
		k = tok.Kind()
	}

	switch k {
	default:
		return &json.SemanticError{JSONKind: k, GoType: cv.Type()}
	case jsontext.NULL:
		if cv.Validator != nil {
			if err := cv.Validate(jsontext.Value("null")); err != nil {
				return WrapPositionError(err, d.StackPointer())
			}
		}
		return cv.Each(d, json.JoinOptions(o...))
	case jsontext.OBJECT:
		s, err := scanner.Structs.Scan(cv.Type())
		if err != nil {
			return err
		}

		var (
			inlined, _ = s.Inlined()
			unknownbuf *bytes.Buffer
			unknownenc *jsontext.Encoder
		)

		if inlined != nil {
			unknownbuf, unknownenc = new(bytes.Buffer), jsontext.NewEncoder(unknownbuf)
			if err = unknownenc.WriteToken(jsontext.BeginObject); err != nil {
				return err
			}
		}

		seen := make(map[string]struct{})
		for d.PeekKind() != jsontext.OBJECT_E {
			var tok jsontext.Token
			tok, err = d.ReadToken()
			if err != nil {
				return err
			}

			name := tok.String()
			seen[name] = struct{}{}

			f, ok := s.Lookup(name)
			if !ok {
				if inlined != nil {
					var val jsontext.Value
					val, err = d.ReadValue()
					if err != nil {
						return err
					}
					if err = unknownenc.WriteToken(jsontext.String(name)); err != nil {
						return err
					}
					if err = unknownenc.WriteValue(val); err != nil {
						return err
					}
				} else {
					if err = d.SkipValue(); err != nil {
						return err
					}
				}
				continue
			}

			var va Validator
			va, err = NewFromStructField(f)
			if err != nil {
				return err
			}

			v := &value{Value: f.GetOrNewAt(cv.Value), Validator: va}
			if err = v.UnmarshalDecode(d, f.PatchOptions(o...)); err != nil {
				return err
			}
		}
		if _, err = d.ReadToken(); err != nil {
			return err
		}

		if inlined != nil {
			if err = unknownenc.WriteToken(jsontext.EndObject); err != nil {
				return err
			}
			v := &value{Value: inlined.GetOrNewAt(cv.Value)}
			if err = v.UnmarshalDecode(jsontext.NewDecoder(unknownbuf), o...); err != nil {
				return err
			}
		}

		if err = cv.each(s, seen, d, o...); err != nil {
			return err
		}

		return nil
	}
}

func (cv *_struct) Each(d *jsontext.Decoder, o json.Options) error {
	s, err := scanner.Structs.Scan(cv.Type())
	if err != nil {
		return err
	}
	return cv.each(s, make(map[string]struct{}), d, o)
}

func (cv *_struct) each(s *scanner.Struct, seen map[string]struct{}, d *jsontext.Decoder, o ...json.Options) error {
	pos := d.StackPointer()

	for f := range s.Range {
		if _, ok := seen[f.Name]; ok {
			continue
		}

		va, err := NewFromStructField(f)
		if err != nil {
			return err
		}

		defaults := []byte("")
		if x, ok := va.(WithDefaults); ok {
			defaults = x.Defaults()
		}
		if len(defaults) == 0 {
			if x, ok := va.(WithOptional); va == nil || ok && x.Optional() {
				continue
			}
			defaults = []byte("null")
		}

		v := &value{Value: f.GetOrNewAt(cv.Value), Validator: va}
		o2 := f.PatchOptions(o...)
		d2 := jsontext.NewDecoder(bytes.NewBuffer(defaults), o2)

		if err = v.UnmarshalDecode(d2, o2); err != nil {
			return WrapPositionError(err, pos+"/"+jsontext.Pointer(f.Name))
		}
	}
	return nil
}

func Map(rv reflect.Value, va Validator) UnmarshalDecoder {
	return &_map{Value: rv, Validator: va}
}

type _map struct {
	reflect.Value
	Validator
}

func (kv *_map) UnmarshalDecode(d *jsontext.Decoder, o ...json.Options) error {
	tok, err := d.ReadToken()
	if err != nil {
		return err
	}

	k := tok.Kind()
	switch k {
	case jsontext.NULL:
		if kv.Validator != nil {
			if err = kv.Validate([]byte("null")); err != nil {
				return WrapPositionError(err, d.StackPointer())
			}
		}
		kv.SetZero()
		return nil
	case jsontext.OBJECT:
		if kv.IsNil() {
			kv.Set(reflect.MakeMap(kv.Type()))
		}

		var (
			kva Validator
			eva Validator
		)
		if x, ok := kv.Validator.(WithKeyRule); ok {
			kva, err = New(Option{Type: kv.Type().Key(), Rule: x.KeyRule()})
			if err != nil {
				return &json.SemanticError{Err: err, GoType: kv.Type()}
			}
		}
		if x, ok := kv.Validator.(WithElemRule); ok {
			eva, err = New(Option{Type: kv.Type().Elem(), Rule: x.ElemRule()})
			if err != nil {
				return &json.SemanticError{Err: err, GoType: kv.Type()}
			}
		}

		var (
			vk = &value{Value: reflect.New(kv.Type().Key()).Elem(), Validator: kva}
			ve = &value{Value: reflect.New(kv.Type().Elem()).Elem(), Validator: eva}
		)

		// NOTE: no need to check duplicate key. JSON decoder will check when ReadValue
		// if jsontext.AllowDuplicateNames is configurated.
		for d.PeekKind() != jsontext.OBJECT_E {
			vk.SetZero()
			if err = vk.UnmarshalDecode(d, o...); err != nil {
				return err
			}
			ve.SetZero()
			if err = ve.UnmarshalDecode(d, o...); err != nil {
				return err
			}
			kv.SetMapIndex(vk.Value, ve.Value)
		}

		if _, err = d.ReadToken(); err != nil {
			return err
		}

		if x, ok := kv.Validator.(PostValidator); ok {
			if err = x.PostValidate(kv.Value); err != nil {
				return err
			}
		}

		return nil
	default:
		return &json.SemanticError{JSONKind: k, GoType: kv.Type()}
	}
}
