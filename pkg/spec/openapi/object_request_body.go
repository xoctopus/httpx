package openapi

import (
	"github.com/xoctopus/httpx/pkg/spec/openapi/jsonschema"
)

type RequestBodyObject struct {
	Description string `json:"description,omitzero"`
	Required    bool   `json:"required,omitzero"`

	ContentObject

	jsonschema.Ext
}
