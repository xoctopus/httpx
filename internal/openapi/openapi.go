package openapi

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/xoctopus/x/docx"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/stringsx"

	"github.com/xoctopus/httpx/internal/method"
	"github.com/xoctopus/httpx/internal/operator"
	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/payload/transformer"
	"github.com/xoctopus/httpx/internal/route"
	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/transport"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
	oas "github.com/xoctopus/httpx/pkg/spec/openapi"
	"github.com/xoctopus/httpx/pkg/spec/openapi/extractors"
	"github.com/xoctopus/httpx/pkg/spec/openapi/jsonschema"
)

type BuildFunc func(r route.Router, fns ...BuildOptionFunc) *oas.OpenAPI

var cached = sync.Map{}

var DefaultBuildFunc = func(r route.Router, fns ...BuildOptionFunc) *oas.OpenAPI {
	if v, ok := cached.Load(r); ok {
		return v.(*oas.OpenAPI)
	}
	o := FromRouter(r, fns...)
	cached.Store(r, o)
	return o
}

type ResponseStatusCodeSpecified interface {
	ResponseStatusCode() int
}

type ResponseContentTypeSpecified interface {
	ResponseContentType() string
}

type ResponseContentSpecified interface {
	ResponseContent() any
}

type ResponseErrorsSpecified interface {
	ResponseErrors() []error
}

func Naming(naming func(t string) string) BuildOptionFunc {
	return func(o *BuildOption) {
		o.naming = naming
	}
}

var gDefaultPkgNamingPrefix = PkgNamingPrefix{}

func RegisterPkgNamingPrefix(pkgPath string, prefix string) {
	gDefaultPkgNamingPrefix.Register(pkgPath, prefix)
}

type PkgNamingPrefix map[string]string

func (p PkgNamingPrefix) Prefix(pkgPath string, name string) string {
	for _, pp := range slices.Sorted(maps.Keys(p)) {
		if strings.HasPrefix(pkgPath, pp) {
			return stringsx.UpperCamelCase(p[pp] + "_" + name)
		}
	}

	return stringsx.UpperCamelCase(name)
}

func (p PkgNamingPrefix) Register(pkgpath string, prefix string) {
	p[pkgpath] = prefix
}

type BuildOptionFunc func(o *BuildOption)

type BuildOption struct {
	naming func(t string) string
}

func FromRouter(r route.Router, options ...BuildOptionFunc) *oas.OpenAPI {
	b := &oass{
		spec:   oas.NewOpenAPI(),
		option: BuildOption{},
	}

	for i := range options {
		options[i](&b.option)
	}

	if b.option.naming == nil {
		naming := func(t string) string {
			var pkgpath string

			splitter := map[string]bool{
				"internal": true,
				"pkg":      true,
				"apis":     true,
				"api":      true,
				"client":   true,
				"domain":   true,
			}

			if i := strings.Index(t, "["); i > 0 {
				base := t[0:i]

				str := &strings.Builder{}

				for k, x := range strings.Split(t[i+1:len(t)-1], ",") {
					if k > 0 {
						str.WriteString("And")
					}
					str.WriteString(b.option.naming(x))
				}

				str.WriteString("As")

				if j := strings.LastIndex(base, "."); j > 0 {
					pkgpath = base[0:j]
					str.WriteString(base[j+1:])
				} else {
					str.WriteString(base)
				}

				return gDefaultPkgNamingPrefix.Prefix(pkgpath, str.String())
			}

			if j := strings.LastIndex(t, "."); j > 0 {
				pkgpath = t[0:j]
			}

			parts := strings.Split(t, "/")

			idx := 0
			for i, p := range parts {
				if splitter[p] {
					idx = i
				}
			}

			if idx < len(parts)-1 {
				t = strings.Join(parts[idx+1:], "/")
			} else {
				t = strings.Join(parts[idx:], "/")
			}

			parts = strings.Split(t, ".")

			if len(parts) == 2 && strings.EqualFold(parts[0], parts[1]) { // strings.ToLower(parts[0]) == strings.ToLower(parts[1]) {
				return gDefaultPkgNamingPrefix.Prefix(pkgpath, parts[0])
			}

			return gDefaultPkgNamingPrefix.Prefix(pkgpath, t)
		}

		b.option.naming = naming
	}

	routes := r.Routes()

	for i := range routes {
		if err := b.scanWithRecover(routes[i]); err != nil {
			panic(fmt.Errorf("failed to scan OpenAPI spec: %w", err))
		}
	}

	return b.spec
}

// oass OpenAPI spec scanner
type oass struct {
	spec     *oas.OpenAPI
	cache    sync.Map
	incoming transport.Incoming
	option   BuildOption
}

func (b *oass) Record(tref string) bool {
	_, ok := b.cache.Load(tref)
	defer b.cache.Store(tref, true)
	return ok
}

func tag(pkgpath string) string {
	tags := strings.Split(pkgpath, "/")
	return tags[len(tags)-1]
}

func (b *oass) scan(r route.Route) error {
	handlers, err := route.NewHandlers(r, "openapi")
	if err != nil {
		return err
	}

	for _, rh := range handlers {
		op := oas.NewOperation(rh.OperationID())

		op.Summary = rh.Summary()
		op.Description = rh.Description()

		if rh.Deprecated() {
			op.Deprecated = new(true)
		}

		ctx := context.Background()

		for _, o := range rh.Operators() {
			b.scanParameterOrRequestBody(ctx, op, o.Type)

			if o.IsLast {
				// FIXME make configurable
				op.Tags = []string{
					tag(o.Type.PkgPath()),
				}

				b.scanResponse(ctx, op, o)
			}

			b.scanResponseError(ctx, op, o)
		}

		b.spec.AddOperation(rh.Method(), b.patchPath(rh.Path(), op), op)
	}

	return nil
}

func (b *oass) scanWithRecover(r route.Route) (err error) {
	defer func() {
		if x := recover(); x != nil {
			switch e := x.(type) {
			case error:
				err = e
			default:
				err = fmt.Errorf("unknown error: %v", x)
			}
		}
	}()

	return b.scan(r)
}

var reHttpRouterPath = regexp.MustCompile("/{([^/]+)(...)?}")

func (b *oass) patchPath(path string, o *oas.OperationObject) string {
	return reHttpRouterPath.ReplaceAllStringFunc(path, func(str string) string {
		name := reHttpRouterPath.FindAllStringSubmatch(str, -1)[0][1]

		// if strings.HasSuffix(name, "...") {
		// 	name = name[0 : len(name)-3]
		// }
		name = strings.TrimSuffix(name, "...")

		isParameterDefined := false

		for _, parameter := range o.Parameters {
			if parameter.In == "path" && parameter.Name == name {
				isParameterDefined = true
			}
		}

		if isParameterDefined {
			return "/{" + name + "}"
		}

		return "/0"
	})
}

func (b *oass) RefString(ref string) string {
	return fmt.Sprintf("#/components/schemas/%s", b.option.naming(ref))
}

func (b *oass) RegisterSchema(ref string, s jsonschema.Schema) {
	if b.spec.ComponentsObject.Schemas == nil {
		b.spec.ComponentsObject.Schemas = map[string]jsonschema.Schema{}
	}

	n := strings.TrimPrefix(ref, "#/components/schemas/")

	if _, ok := b.spec.ComponentsObject.Schemas[n]; !ok {
		b.spec.ComponentsObject.Schemas[n] = s
	} else {
		fmt.Println(n, "Registered.")
	}
}

func (b *oass) SchemaFromType(ctx context.Context, v any, def bool) jsonschema.Schema {
	return extractors.SchemaFrom(extractors.WithSchemaRegister(ctx, b), v, def)
}

func (b *oass) scanResponse(ctx context.Context, op *oas.OperationObject, o *operator.Factory) {
	mtd := ""

	var (
		statusCode  = http.StatusNoContent
		contentType = "application/json"
		resp        = &oas.ResponseObject{}
	)

	if can, ok := o.Operator.(method.Describer); ok {
		mtd = can.Method()

		if mtd == http.MethodPost {
			statusCode = http.StatusCreated
		} else {
			statusCode = http.StatusOK
		}
	}

	if mtd == "" {
		return
	}

	if x, ok := o.Operator.(ResponseStatusCodeSpecified); ok {
		statusCode = x.ResponseStatusCode()
	}

	if x, ok := o.Operator.(ResponseContentTypeSpecified); ok {
		contentType = x.ResponseContentType()
	}

	if x, ok := o.Operator.(ResponseContentSpecified); ok {
		if rt := x.ResponseContent(); rt != nil {
			if c, ok := rt.(content.MediaTypeDescriber); ok {
				contentType = c.ContentType()
			}

			mt := &oas.MediaTypeObject{}
			mt.Schema = b.SchemaFromType(ctx, rt, false)
			resp.AddContent(contentType, mt)
		} else {
			statusCode = http.StatusNoContent
		}
	} else {
		resp.AddContent(contentType, &oas.MediaTypeObject{})
	}

	op.AddResponse(statusCode, resp)
}

func (b *oass) scanResponseError(ctx context.Context, op *oas.OperationObject, o *operator.Factory) {
	if can, ok := o.Operator.(ResponseErrorsSpecified); ok {
		returnErrors := can.ResponseErrors()

		codes := map[int][]string{}

		for _, err := range returnErrors {
			if e, ok := err.(status.Describer); ok {
				codes[e.StatusCode()] = append(codes[e.StatusCode()], err.Error())
			}
		}

		if op.Responses == nil {
			op.Responses = map[string]*oas.ResponseObject{}
		}

		for statusCode := range codes {
			errResp, ok := op.Responses[fmt.Sprintf("%d", statusCode)]
			if !ok {
				errResp = &oas.ResponseObject{}
			}

			mt := &oas.MediaTypeObject{}

			if e, ok := errors.AsType[*status.Description](returnErrors[0]); ok {
				mt.Schema = b.SchemaFromType(ctx, &status.Response{}, false)
			} else {
				mt.Schema = b.SchemaFromType(ctx, e, false)
			}

			errResp.AddContent("application/json", mt)

			if found, ok := errResp.GetExtension("x-status-return-errors"); ok {
				errResp.AddExtension("x-status-return-errors", append(found.([]string), codes[statusCode]...))
			} else {
				errResp.AddExtension("x-status-return-errors", codes[statusCode])
			}

			op.AddResponse(statusCode, errResp)
		}
	}
}

func (b *oass) scanParameterOrRequestBody(ctx context.Context, op *oas.OperationObject, t reflect.Type) {
	var dp docx.Provider

	if d, ok := reflect.New(t).Interface().(docx.Provider); ok {
		dp = d
	}

	s, err := scanner.Structs.Scan(t)
	must.NoErrorF(err, "failed to scan %s", t)

	for f := range s.Range {
		location := f.Tag.Get("in")
		must.BeTrueF(
			len(location) > 0,
			"missing location operation:%s field:%s",
			op.OperationId, f.FieldName,
		)
		optional := f.Omitzero || f.Omitempty

		tf, err := transformer.New(f.Type, f.Tag.Get("mime"), transformer.ForUnmarshalling)
		must.NoErrorF(
			err,
			"failed to new transformer operation:%s field:%s",
			op.OperationId, f.FieldName,
		)

		schema := b.SchemaFromType(ctx, reflect.New(f.Type).Interface(), false)
		if schema != nil {
			option := validation.Option{Type: f.Type}
			if text := f.Tag.Get("validate"); len(text) > 0 {
				option.Rule = rule.MustCompile(text)
			}
			patched, err := extractors.PatchSchemaValidation(schema, option)
			must.NoErrorF(
				err,
				"failed to patch validation operation:%s field:%s",
				op.OperationId, f.FieldName,
			)
			schema = patched
		}

		if schema != nil && dp != nil {
			if lines, ok := dp.DocOf(f.FieldName); ok {
				extractors.SetTitleOrDescription(schema.GetMetadata(), lines)
			}
		}

		switch location {
		case "body":
			reqBody := op.RequestBody
			if op.RequestBody == nil {
				reqBody = &oas.RequestBodyObject{Required: true}
				op.SetRequestBody(reqBody)
			}
			reqBody.AddContent(tf.Media(), &oas.MediaTypeObject{Schema: schema})
		case "query":
			op.AddParameter(f.Name, oas.InQuery, &oas.Parameter{
				Schema:   schema,
				Required: new(!optional),
			})
		case "cookie":
			op.AddParameter(f.Name, oas.InCookie, &oas.Parameter{
				Schema:   schema,
				Required: new(!optional),
			})
		case "header":
			op.AddParameter(f.Name, oas.InHeader, &oas.Parameter{
				Schema:   schema,
				Required: new(!optional),
			})
		case "path":
			op.AddParameter(f.Name, oas.InPath, &oas.Parameter{
				Schema:   schema,
				Required: new(true),
			})
		}
	}
}
