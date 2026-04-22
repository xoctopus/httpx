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

func Unmarshal(data []byte, out any, o ...json.Options) error {
	return UnmarshalRead(bytes.NewReader(data), out, o...)
}

func UnmarshalRead(r io.Reader, out any, o ...json.Options) error {
	return UnmarshalDecode(jsontext.NewDecoder(r), out, o...)
}

func UnmarshalDecode(d *jsontext.Decoder, out any, o ...json.Options) error {
	return validation.UnmarshalDecode(d, out, o...)
}

func Marshal(out any, o ...json.Options) ([]byte, error) {
	return json.Marshal(out, json.MigrateOmitzero(o)...)
}

func MarshalWrite(w io.Writer, out any, o ...json.Options) error {
	return json.MarshalWrite(w, out, json.MigrateOmitzero(o)...)
}

func MarshalEncode(e *jsontext.Encoder, out any, o ...json.Options) error {
	return json.MarshalEncode(e, out, json.MigrateOmitzero(o)...)
}
