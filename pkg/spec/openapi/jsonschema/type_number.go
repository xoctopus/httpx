package jsonschema

import (
	"io"
)

func Integer() Schema {
	t := &NumberType{
		Type:    "integer",
		Minimum: new(float64(-1 << (32 - 1))),
		Maximum: new(float64(1<<(32-1) - 1)),
	}

	t.AddExtension("x-format", "int32")
	return t
}

func Long() Schema {
	t := &NumberType{
		Type: "integer",
	}
	t.AddExtension("x-format", "int64")
	return t
}

type NumberType struct {
	Type string `json:"type,omitzero"`

	// validate
	MultipleOf       *float64 `json:"multipleOf,omitzero"`
	Maximum          *float64 `json:"maximum,omitzero"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitzero"`
	Minimum          *float64 `json:"minimum,omitzero"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitzero"`

	Core
	Metadata
}

func (t NumberType) PrintTo(w io.Writer, optionFns ...SchemaPrintOption) {
	Print(w, func(p Printer) {
		if t.Type == "integer" {
			p.Print("int")
			return
		}
		p.Print("number")
	})
}
