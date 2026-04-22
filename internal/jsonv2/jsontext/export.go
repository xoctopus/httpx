package jsontext

import (
	"github.com/go-json-experiment/json/jsontext"
)

type (
	Value          = jsontext.Value
	Pointer        = jsontext.Pointer
	Token          = jsontext.Token
	Kind           = jsontext.Kind
	Encoder        = jsontext.Encoder
	Decoder        = jsontext.Decoder
	SyntacticError = jsontext.SyntacticError
)

var (
	NewEncoder          = jsontext.NewEncoder
	NewDecoder          = jsontext.NewDecoder
	AllowDuplicateNames = jsontext.AllowDuplicateNames
)

func AppendQuote[T ~[]byte | ~string](dst []byte, src T) ([]byte, error) {
	return jsontext.AppendQuote(dst, src)
}

func AppendUnquote[T ~[]byte | ~string](dst []byte, src T) ([]byte, error) {
	return jsontext.AppendUnquote(dst, src)
}
