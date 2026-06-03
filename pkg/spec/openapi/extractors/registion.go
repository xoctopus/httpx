package extractors

import (
	"context"
	"sync"

	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/httpx/pkg/spec/openapi/jsonschema"
)

type register struct {
	m sync.Map
}

var gDefaultRegister = &register{}

func (d *register) Record(tref string) bool {
	_, ok := d.m.Load(tref)
	defer d.m.Store(tref, true)
	return ok
}

func (d *register) RegisterSchema(ref string, s jsonschema.Schema) {}

func (d *register) RefString(ref string) string {
	return ref
}

type SchemaRegister interface {
	RegisterSchema(ref string, s jsonschema.Schema)
	RefString(ref string) string
	Record(typeRef string) bool
}

type tCtxRegister struct{}

var (
	WithSchemaRegister  = contextx.With[tCtxRegister, SchemaRegister]
	MustSchemaRegister  = contextx.Must[tCtxRegister, SchemaRegister]
	CarrySchemaRegister = contextx.Carry[tCtxRegister, SchemaRegister]
)

func SchemaRegisterFrom(ctx context.Context) SchemaRegister {
	r, ok := contextx.From[tCtxRegister, SchemaRegister](ctx)
	if ok {
		return r
	}
	return gDefaultRegister
}
