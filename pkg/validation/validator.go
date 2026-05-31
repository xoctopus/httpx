package validation

import (
	"bytes"
	"io"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
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
	return json.Marshal(out, json.MigrateOmitzero(o)...)
}

func MarshalWrite(w io.Writer, v any, o ...json.Options) error {
	return json.MarshalWrite(w, v, json.MigrateOmitzero(o)...)
}

func MarshalEncode(e *jsontext.Encoder, v any, o ...json.Options) error {
	return json.MarshalEncode(e, v, json.MigrateOmitzero(o)...)
}
