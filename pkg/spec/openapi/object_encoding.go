package openapi

import (
	"github.com/xoctopus/httpx/pkg/spec/openapi/jsonschema"
)

type EncodingObject struct {
	ContentType string `json:"contentType"`

	HeadersObject

	Style         ParameterStyle `json:"style,omitzero"`
	Explode       bool           `json:"explode,omitzero"`
	AllowReserved bool           `json:"allowReserved,omitzero"`

	jsonschema.Ext
}
