package jsonschema

import (
	"bytes"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
)

type OpenAPISchemaGetter interface {
	OpenAPISchema() Schema
}

type OpenAPISchemaTypeGetter interface {
	OpenAPISchemaType() []string
}

type OpenAPISchemaFormatGetter interface {
	OpenAPISchemaFormat() string
}

// CanSwaggerDoc interface of k8s pkgs
type CanSwaggerDoc interface {
	SwaggerDoc() map[string]string
}

type Payload struct {
	Schema
}

func Unmarshal(data []byte, v any) error {
	if err := json.UnmarshalDecode(
		jsontext.NewDecoder(bytes.NewReader(data)),
		v,
		json.WithUnmarshalers(unmarshalers),
	); err != nil {
		return err
	}
	return nil
}

func (p Payload) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Schema)
}

func (p *Payload) UnmarshalJSON(data []byte) (err error) {
	var schema Schema
	if err = json.UnmarshalDecode(
		jsontext.NewDecoder(bytes.NewReader(data)),
		&schema,
		json.WithUnmarshalers(unmarshalers),
	); err != nil {
		return err
	}
	*p = Payload{
		Schema: schema,
	}
	return nil
}
