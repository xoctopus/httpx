package validation

import (
	"bytes"
	jsonv1 "encoding/json"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io"

	"github.com/xoctopus/httpx/internal/validation"
	_ "github.com/xoctopus/httpx/pkg/validation/regex"
)

type (
	Provider  = validation.Provider
	Validator = validation.Validator
)

func Unmarshal(data []byte, v any, o ...json.Options) error {
	return UnmarshalReader(bytes.NewReader(data), v, o...)
}

func UnmarshalReader(r io.Reader, v any, o ...json.Options) error {
	return UnmarshalDecode(jsontext.NewDecoder(r), v, o...)
}

func UnmarshalDecode(d *jsontext.Decoder, v any, o ...json.Options) error {
	return validation.UnmarshalDecode(d, v, o...)
}

func Marshal(out any, o ...json.Options) ([]byte, error) {
	return json.Marshal(out, append(o, jsonv1.OmitEmptyWithLegacySemantics(true))...)
}

func MarshalWrite(w io.Writer, v any, o ...json.Options) error {
	return json.MarshalWrite(w, v, append(o, jsonv1.OmitEmptyWithLegacySemantics(true))...)
}

func MarshalEncode(e *jsontext.Encoder, v any, o ...json.Options) error {
	return json.MarshalEncode(e, v, append(o, jsonv1.OmitEmptyWithLegacySemantics(true))...)
}
