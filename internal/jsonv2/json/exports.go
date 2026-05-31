// Package json exports defines from github.com/go-json-experiment/json/jsontext
// after golang release json/v2 use json/v2 instead

package json

import (
	"github.com/go-json-experiment/json"
	jsonv1 "github.com/go-json-experiment/json/v1"
)

type (
	Marshaler       = json.Marshaler
	MarshalerTo     = json.MarshalerTo
	Unmarshaler     = json.Unmarshaler
	UnmarshalerFrom = json.UnmarshalerFrom
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

	GetOption        = json.GetOption[bool]
	JoinOptions      = json.JoinOptions
	StringifyNumbers = json.StringifyNumbers
)

func MigrateOmitzero(o []Options) []Options {
	return append(o, jsonv1.OmitEmptyWithLegacySemantics(true))
}
