package openapi

import (
	"net/url"
	"strings"

	"github.com/xoctopus/httpx/pkg/spec/openapi/jsonschema"
)

type ComponentsObject struct {
	Schemas map[string]jsonschema.Schema `json:"schemas,omitzero"`
}

func (o *ComponentsObject) AddSchema(id string, s jsonschema.Schema) {
	if s == nil {
		return
	}
	if o.Schemas == nil {
		o.Schemas = make(map[string]jsonschema.Schema)
	}
	o.Schemas[id] = s
}

func (o *ComponentsObject) RefSchema(id string) jsonschema.Schema {
	if o.Schemas == nil || o.Schemas[id] == nil {
		return nil
	}
	ref := url.URL{Fragment: strings.Join([]string{"#", "components", "schemas", id}, "/")}
	return &jsonschema.RefType{Ref: new(jsonschema.URIRef(ref))}
}
