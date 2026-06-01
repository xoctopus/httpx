package openapi

import (
	"github.com/xoctopus/httpx/pkg/spec/openapi/internal"
)

// PathItemObject
// https://spec.openapis.org/oas/latest.html#pathItemObject
// no need $ref, server, parameters
type PathItemObject = internal.Record[string, *OperationObject]

type CallbacksObject struct {
	Callbacks internal.Record[string, *PathItemObject] `json:"callbacks,omitzero"`
}

func (o *CallbacksObject) AddCallback(name string, c *PathItemObject) {
	o.Callbacks.Set(name, c)
}
