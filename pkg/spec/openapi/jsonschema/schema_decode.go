package jsonschema

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
)

type Schema interface {
	GetCore() *Core
	GetMetadata() *Metadata
	PrintTo(w io.Writer, options ...SchemaPrintOption)
}

type ValidationSchema interface {
	PatchValidation(any) error
}

var (
	unmarshalers = json.UnmarshalFromFunc[*Schema](
		func(d *jsontext.Decoder, schema *Schema) error {
			return (&decoder{schema: schema}).UnmarshalJSONFrom(d)
		},
	)
)

var (
	ErrInvalidJSONSchemaObject = errors.New("invalid json schema object")
	ErrInvalidJSONSchemaType   = errors.New("invalid json schema type")
)

type decoder struct {
	schema  *Schema
	options json.Options
	// anchors map[string]string
}

var _ json.UnmarshalerFrom = &decoder{}

func (u *decoder) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	u.options = d.Options()

	startToken, err := d.ReadToken()
	if err != nil {
		return err
	}

	switch startToken.Kind() {
	case jsontext.TRUE:
		*u.schema = &AnyType{}
		return nil
	case jsontext.OBJECT:
		return u.unmarshal(d)
	}

	return ErrInvalidJSONSchemaObject
}

func (u *decoder) decode(d *jsontext.Decoder, target any) error {
	k, err := d.ReadValue()
	if err != nil {
		return err
	}
	if err = json.Unmarshal(k, target, u.options); err != nil {
		return err
	}
	return nil
}

func (u *decoder) new(typ string, format string) (Schema, error) {
	switch typ {
	case "array":
		return &ArrayType{Type: typ}, nil
	case "object":
		return &ObjectType{Type: typ}, nil
	case "number":
		t := &NumberType{Type: "number"}
		switch format {
		case "int64", "int32", "int16", "int8", "uint64", "uint32", "uint16", "uint8":
			t.AddExtension("x-format", format)
		case "float":
			t.AddExtension("x-format", "float32")
		case "double":
			t.AddExtension("x-format", "float64")
		}
		return t, nil
	case "integer", "int":
		t := &NumberType{Type: "integer"}
		switch format {
		case "int64", "int32", "int16", "int8", "uint64", "uint32", "uint16", "uint8":
			t.AddExtension("x-format", format)
		}
		return t, nil
	case "string":
		return &StringType{Type: typ}, nil
	case "null":
		return &NullType{Type: typ}, nil
	case "boolean":
		return &BooleanType{Type: typ}, nil
	}
	return nil, ErrInvalidJSONSchemaType
}

func (u *decoder) unmarshal(d *jsontext.Decoder) error {
	unprocessed := bytes.NewBuffer(nil)
	unprocessedenc := jsontext.NewEncoder(unprocessed)

	_ = unprocessedenc.WriteToken(jsontext.BeginObject)

	var (
		schema    any
		typ       string
		format    string
		additions []Schema
	)

	for kind := d.PeekKind(); kind != jsontext.OBJECT_E; kind = d.PeekKind() {
		var prop string
		if err := u.decode(d, &prop); err != nil {
			return fmt.Errorf("decode prop failed: %w", err)
		}

		// renaming
		switch prop {
		case "$recursiveRef":
			prop = "$dynamicRef"
		case "$recursiveAnchor":
			prop = "$dynamicAnchor"
		case "definitions":
			prop = "$def"
		case "dependencies":
			// TODO convert to with dependentSchemas and dependentRequired
		}

		switch prop {
		case "const":
			var value any
			if err := u.decode(d, &value); err != nil {
				return fmt.Errorf("decode prop %s failed: %w", prop, err)
			}
			schema = &EnumType{Enum: []any{value}}
			continue // skip unmarshal decode const
		case "format":
			if err := u.decode(d, &format); err != nil {
				return fmt.Errorf("decode prop %s failed: %w", prop, err)
			}
			continue
		case "enum":
			schema = &EnumType{}
		case "items", "prefixItems":
			schema = &ArrayType{Type: "array"}
		case "properties", "propertyNames", "patternProperties", "additionalProperties", "required":
			schema = &ObjectType{Type: "object"}
		case "oneOf", "discriminator":
			schema = &UnionType{}
		case "allOf":
			schema = &IntersectionType{}
		case "$dynamicRef":
			schema = &RefType{}
		case "$ref":
			schema = &RefType{}
		case "type":
			v, err := d.ReadValue()
			if err != nil {
				return err
			}
			switch v.Kind() {
			case jsontext.ARRAY:
				var union []string
				if err = json.Unmarshal(v, &union); err != nil {
					return err
				}
				if len(union) > 0 {
					typ = union[0]
				}
				for i, t := range union {
					if i == 0 {
						typ = t
						continue
					}
					s, err := u.new(t, "")
					if err != nil {
						return err
					}
					additions = append(additions, s)
				}
				continue
			default:
				if err = json.Unmarshal(v, &typ); err != nil {
					return err
				}
			}
			continue
		}

		v, err := d.ReadValue()
		if err != nil {
			return fmt.Errorf("read prop %s failed: %w", prop, err)
		}
		_ = json.MarshalEncode(unprocessedenc, prop)
		_ = json.MarshalEncode(unprocessedenc, v)
	}

	// read the EndObject to mark dec finished
	t, err := d.ReadToken()
	if err != nil {
		return err
	}
	_ = unprocessedenc.WriteToken(t)

	if schema == nil || len(additions) == 0 {
		if typ != "" {
			s, err := u.new(typ, format)
			if err != nil {
				return err
			}
			schema = s
		}
	}

	if schema == nil {
		schema = &AnyType{}
	}

	// {}\n
	if unprocessed.Len() > 3 {
		if err = json.UnmarshalRead(unprocessed, schema, u.options); err != nil {
			return err
		}
	}

	if it, ok := schema.(*IntersectionType); ok {
		// TODO for old structure
		if len(it.AllOf) == 2 {
			switch x := it.AllOf[1].(type) {
			case *ObjectType:
				// skip
			default:
				s := it.AllOf[0]
				x.GetMetadata().DeepCopyInto(s.GetMetadata())
				schema = s
			}
		}
	}

	if len(additions) > 0 {
		*u.schema = OneOf(
			append([]Schema{
				schema.(Schema),
			}, additions...)...,
		)
	} else {
		*u.schema = schema.(Schema)
	}

	return nil
}
