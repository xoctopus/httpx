package validators

import (
	"errors"
	"unicode/utf8"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/internal/validation/va"
)

func init() {
	validation.Register(&_stringP{})
}

type _stringP struct{}

func (_stringP) Name() string {
	return "string"
}

func (_stringP) Variants() []string {
	return []string{"string"}
}

func (p *_stringP) New(r rule.Rule) (_ validation.Validator, err error) {
	v := &String{mode: BYTE_MODE}

	if v.LengthVa, err = va.NewLengthVa(r); err != nil {
		return nil, err
	}
	if v.EnumVa, err = va.NewEnumVa[string](r); err != nil {
		return nil, err
	}
	v.RegexpVa = va.NewRegexpVa(r, "")

	params := r.Parameters()
	if len(params) == 0 {
		return v, nil
	}
	if len(params) != 1 {
		return nil, codex.Errorf(validation.ERROR__STRING_LENGTH_MODE, "most 1 parameter")
	}
	mode := string(params[0].Bytes())
	if mode != BYTE_MODE && mode != RUNE_MODE {
		return nil, codex.Errorf(validation.ERROR__STRING_LENGTH_MODE, "got %s", mode)
	}

	return v, nil
}

const (
	BYTE_MODE = "byte"
	RUNE_MODE = "rune"
)

var LenFunc = map[string]func(string) uint{
	BYTE_MODE: func(s string) uint { return uint(len(s)) },
	RUNE_MODE: func(s string) uint { return uint(utf8.RuneCount([]byte(s))) },
}

type String struct {
	*va.LengthVa
	*va.RegexpVa
	*va.EnumVa[string]

	mode string
}

func (v *String) Validate(value []byte) error {
	if k := jsontext.Value(value).Kind(); k != jsontext.STRING {
		return codex.Errorf(validation.ERROR__INPUT_TYPE, "expect: string got: %s", k)
	}

	unquote, err := jsontext.AppendUnquote(nil, value)
	if err != nil {
		return err
	}
	val := string(unquote)

	length := LenFunc[v.mode](val)

	return errors.Join(
		v.LengthVa.Validate(length),
		v.EnumVa.Validate(val),
		v.RegexpVa.Validate(val),
	)
}

func (v *String) String() string {
	b := rule.NewBuilder("string")

	b.AddParameters(rule.NewLiteral(v.mode))

	v.EnumVa.BuiltTo(b)
	v.LengthVa.BuiltTo(b)
	v.RegexpVa.BuiltTo(b)

	return string(b.Bytes())
}
