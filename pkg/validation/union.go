package validation

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/go-json-experiment/json/jsontext"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/validation"
)

type UnionType interface {
	Discriminator() string
	Mapping() map[string]any
	SetUnderlying(u any)
}

func UnmarshalUnion(data []byte, ut UnionType) error {
	return UnmarshalDecode(jsontext.NewDecoder(bytes.NewBuffer(data)), ut)
}

func UnmarshalDecodeUnion(dec *jsontext.Decoder, ut UnionType) error {
	t, err := dec.ReadToken()
	if err != nil {
		return err
	}

	switch t.Kind() {
	case jsontext.KindNull:
		return nil
	case jsontext.KindBeginObject:
		discriminatorValue := ""

		buf := bytes.NewBuffer(nil)
		enc := jsontext.NewEncoder(buf)
		if err := enc.WriteToken(jsontext.BeginObject); err != nil {
			return err
		}

		for dec.PeekKind() != jsontext.KindEndObject {
			k, err := dec.ReadToken()
			if err != nil {
				return err
			}
			propName := k.String()

			propValue, err := dec.ReadValue()
			if err != nil {
				return err
			}

			if propName == ut.Discriminator() {
				if err := Unmarshal(propValue, &discriminatorValue); err != nil {
					return err
				}
			}

			if err := enc.WriteToken(jsontext.String(propName)); err != nil {
				return err
			}
			if err := enc.WriteValue(propValue); err != nil {
				return err
			}
		}

		// read }
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		if err := enc.WriteToken(jsontext.EndObject); err != nil {
			return err
		}

		if v, ok := ut.Mapping()[discriminatorValue]; ok {
			if err := UnmarshalDecode(jsontext.NewDecoder(buf), v); err != nil {
				return validation.WrapPositionError(err, dec.StackPointer())
			}
			ut.SetUnderlying(v)
			return nil
		}

		if discriminatorValue == "" {
			// when empty discriminatorValue should drop other fields
			return nil
		}

		return validation.WrapPositionError(
			fmt.Errorf("unsupported %s=%s", ut.Discriminator(), discriminatorValue),
			jsontext.Pointer(fmt.Sprintf("/%s", ut.Discriminator())),
		)
	default:
		return &json.SemanticError{Err: errors.New("expect object or null")}
	}
}
