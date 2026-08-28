package jsontext

import (
	"encoding/json/jsontext"
)

// Types
type (
	Value          = jsontext.Value
	Pointer        = jsontext.Pointer
	Token          = jsontext.Token
	Kind           = jsontext.Kind
	Encoder        = jsontext.Encoder
	Decoder        = jsontext.Decoder
	SyntacticError = jsontext.SyntacticError
)

// APIs
var (
	NewEncoder = jsontext.NewEncoder
	NewDecoder = jsontext.NewDecoder
)

// Options
var (
	AllowDuplicateNames = jsontext.AllowDuplicateNames
	_                   = jsontext.CanonicalizeRawFloats
	_                   = jsontext.CanonicalizeRawInts
)

// Token functions
var (
	String  = jsontext.String
	Int     = jsontext.Int
	Uint    = jsontext.Uint
	Float   = jsontext.Float
	Float32 = jsontext.Float32
	Bool    = jsontext.Bool
)

// JSON token kinds
const (
	NULL     = jsontext.KindNull
	NUMBER   = jsontext.KindNumber
	STRING   = jsontext.KindString
	FALSE    = jsontext.KindFalse
	TRUE     = jsontext.KindTrue
	OBJECT   = jsontext.KindBeginObject
	OBJECT_E = jsontext.KindEndObject
	ARRAY    = jsontext.KindBeginArray
	ARRAY_E  = jsontext.KindEndArray
)

// Errors define in jsontext
var (
	ErrDuplicateName = jsontext.ErrDuplicateName
	ErrNonStringName = jsontext.ErrNonStringName
)

func AppendQuote[T ~[]byte | ~string](dst []byte, src T) ([]byte, error) {
	return jsontext.AppendQuote(dst, src)
}

func AppendUnquote[T ~[]byte | ~string](dst []byte, src T) ([]byte, error) {
	return jsontext.AppendUnquote(dst, src)
}

// Tokens
var (
	Null        = jsontext.Null
	False       = jsontext.False
	True        = jsontext.True
	BeginObject = jsontext.BeginObject
	EndObject   = jsontext.EndObject
	BeginArray  = jsontext.BeginArray
	EndArray    = jsontext.EndArray
)
