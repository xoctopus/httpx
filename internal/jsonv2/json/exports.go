// Package json exports defines from github.com/go-json-experiment/json/jsontext
// after golang release json/v2 use json/v2 instead

package json

import (
	jsonv1 "encoding/json"

	"encoding/json/v2"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
)

type (
	Marshaler       = json.Marshaler
	MarshalerTo     = json.MarshalerTo
	Unmarshaler     = json.Unmarshaler
	UnmarshalerFrom = json.UnmarshalerFrom
	Unmarshalers    = json.Unmarshalers
	Options         = json.Options

	SemanticError = json.SemanticError
	SyntaxError   = jsonv1.SyntaxError
)

var (
	Marshal         = json.Marshal
	MarshalEncode   = json.MarshalEncode
	MarshalWrite    = json.MarshalWrite
	Unmarshal       = json.Unmarshal
	UnmarshalDecode = json.UnmarshalDecode
	UnmarshalRead   = json.UnmarshalRead

	GetOption        = json.GetOption[bool]
	JoinOptions      = json.JoinOptions
	StringifyNumbers = json.StringifyNumbers
	WithUnmarshalers = json.WithUnmarshalers
)

func MigrateOmitzero(o []Options) []Options {
	return append(o, jsonv1.OmitEmptyWithLegacySemantics(true))
}

func UnmarshalFromFunc[T any](fn func(*jsontext.Decoder, T) error) *Unmarshalers {
	return json.UnmarshalFromFunc(fn)
}
