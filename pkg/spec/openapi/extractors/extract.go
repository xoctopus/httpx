package extractors

import (
	"context"
	"encoding"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/xoctopus/x/docx"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/reflectx"

	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/pkg/spec/openapi/jsonschema"
)

func SchemaFrom(ctx context.Context, v any, def bool) jsonschema.Schema {
	if v == nil {
		return nil
	}

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return SchemaFromType(ctx, t, Opt{Decl: def})
}

//nolint:gocyclo
func SchemaFromType(ctx context.Context, t reflect.Type, opt Opt) (s jsonschema.Schema) {
	sr := SchemaRegisterFrom(ctx)

	fill := func(tref string) {
		if _, ok := s.(*jsonschema.RefType); !ok {
			if x, ok := reflect.New(t).Interface().(validation.TagValidator); ok {
				if text := x.ValidationTag(); len(text) > 0 {
					p, err := PatchSchemaValidation(s, validation.Option{
						Type: t,
						Rule: rule.MustCompile(text),
					})
					must.NoErrorF(err, "failed to patch `%s` for type %s", text, t)
					s = p
				}
			}
		}

		if s != nil {
			if !(strings.Contains(tref, "/internal/") || strings.Contains(tref, "/internal.")) {
				s.GetMetadata().AddExtension(jsonschema.XGoVendorType, tref)
			}
		}
	}
	inst := reflect.New(t).Interface()

	// named type
	if pkgpath := t.PkgPath(); pkgpath != "" {
		tref := fmt.Sprintf("%s.%s", pkgpath, t.Name())

		defer fill(tref)

		ref := sr.RefString(tref)

		if ok := sr.Record(tref); ok {
			u, err := jsonschema.ParseURIRef(ref)
			must.NoErrorF(err, "failed to parse URIRef from `%s` for type %s", ref, t)
			return &jsonschema.RefType{Ref: u}
		}

		defer func() {
			if n := len(opt.EnumInDoc); n > 0 {
				e := &jsonschema.EnumType{}

				e.Enum = make([]any, n)
				for i := range e.Enum {
					e.Enum[i] = opt.EnumInDoc[i]
				}

				if s != nil {
					s.GetMetadata().DeepCopyInto(e.GetMetadata())
				}
				s = e
			}

			sr.RegisterSchema(ref, s)

			if !opt.Decl {
				u, err := jsonschema.ParseURIRef(ref)
				must.NoErrorF(err, "failed to parse URIRef from `%s` for type %s", ref, t)
				s = &jsonschema.RefType{Ref: u}
			}
		}()

		if x, ok := inst.(jsonschema.CanSwaggerDoc); ok {
			opt = opt.WithDoc(x.SwaggerDoc())
		}

		if x, ok := inst.(jsonschema.GoEnumValues); ok {
			defer func() {
				values := x.EnumValues()
				labels := make([]string, 0)
				e := &jsonschema.EnumType{}

				for i := range values {
					e.Enum = append(e.Enum, values[i])
					// enum label
					if el, ok := values[i].(interface{ Text() string }); ok {
						labels = append(labels, el.Text())
					}
				}
				if len(labels) > 0 {
					e.AddExtension(jsonschema.XEnumLabels, labels)
				}
				if s != nil {
					s.GetMetadata().DeepCopyInto(e.GetMetadata())
				}
				s = e
			}()
		}

		if dp, ok := inst.(docx.Provider); ok {
			ctx = docx.WithProvider(ctx, dp)
			defer func() {
				if lines, ok := dp.DocOf(); ok {
					SetTitleOrDescription(s.GetMetadata(), lines)
				}
			}()
		}

		defer fill(tref)

		if g, ok := inst.(jsonschema.GoUnionType); ok {
			if types := g.OneOf(); len(types) != 0 {
				schemas := make([]jsonschema.Schema, len(types))
				for i := range schemas {
					schemas[i] = SchemaFromType(ctx, reflectx.Deref(reflect.TypeOf(types[i])), opt.WithDecl(false))
				}
				if len(schemas) == 1 {
					return schemas[0]
				}
				return jsonschema.OneOf(schemas...)
			}
		}

		if g, ok := inst.(jsonschema.GoTaggedUnionType); ok {
			types := g.Mapping()
			schemas := make([]jsonschema.Schema, 0, len(types))
			mapping := map[string]jsonschema.Schema{}
			for _, tag := range slices.Sorted(maps.Keys(types)) {
				s := SchemaFromType(
					ctx,
					reflectx.Deref(reflect.TypeOf(types[tag])),
					opt.WithDecl(false),
				)
				schemas = append(schemas, s)
				mapping[tag] = s
			}

			s := jsonschema.OneOf(schemas...)
			s.Discriminator = &jsonschema.Discriminator{
				PropertyName: g.Discriminator(),
				Mapping:      mapping,
			}
			return s
		}

		if g, ok := inst.(jsonschema.OpenAPISchemaFormatGetter); ok {
			s := jsonschema.String()
			s.Format = g.OpenAPISchemaFormat()
			switch s.Format {
			case "int-or-string":
				return jsonschema.OneOf(jsonschema.Integer(), jsonschema.String())
			}
			return s
		}

		if g, ok := inst.(jsonschema.OpenAPISchemaTypeGetter); ok {
			typ := g.OpenAPISchemaType()
			if len(typ) > 0 && typ[0] != "" {
				p := jsonschema.Payload{}

				_ = p.UnmarshalJSON(fmt.Appendf(nil, `{"type":%q}`, typ[0]))

				if p.Schema != nil {
					return p.Schema
				}
			}
		}

		if g, ok := inst.(jsonschema.OpenAPISchemaGetter); ok {
			s := g.OpenAPISchema()
			return s
		}

		// TODO find better way
		if tref == "mime/multipart.FileHeader" || tref == "io.ReadCloser" {
			return jsonschema.Binary()
		}

		if _, ok := inst.(encoding.TextUnmarshaler); ok {
			if _, ok := inst.(encoding.TextMarshaler); ok {
				return jsonschema.String()
			}
		}
	}

	switch t.Kind() {
	case reflect.Pointer:
		count := 1
		elem := t.Elem()

		for elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
			count++
		}
		return func(s jsonschema.Schema) jsonschema.Schema {
			s.GetMetadata().AddExtension(jsonschema.XGoStarLevel, count)
			return s
		}(SchemaFromType(ctx, elem, opt.WithDecl(false)))
	case reflect.Interface:
		return jsonschema.Any()
	case reflect.String:
		return jsonschema.String()
	case reflect.Bool:
		return jsonschema.Boolean()
	case reflect.Float32:
		st := &jsonschema.NumberType{Type: "number"}
		st.AddExtension("x-format", "float32")
		return st
	case reflect.Float64:
		st := &jsonschema.NumberType{Type: "number"}
		st.AddExtension("x-format", "float64")
		return st
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		st := &jsonschema.NumberType{Type: "integer"}
		st.AddExtension("x-format", t.Kind().String())
		return st
	case reflect.Array:
		s := jsonschema.ArrayOf(SchemaFromType(ctx, t.Elem(), opt.WithDecl(false)))
		n := uint64(t.Len())
		s.MaxItems, s.MinItems = new(n), new(n)
		return s
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 && t.Elem().PkgPath() == "" {
			return jsonschema.Bytes()
		}
		item := SchemaFromType(ctx, t.Elem(), opt.WithDecl(false))
		if item == nil {
			item = jsonschema.Any()
		}
		return jsonschema.ArrayOf(item)
	case reflect.Map:
		key := SchemaFromType(ctx, t.Key(), opt.WithDecl(false))
		switch key.(type) {
		case *jsonschema.StringType:
			break
		case *jsonschema.RefType:
			break
		default:
			_, ok := key.(*jsonschema.StringType)
			must.BeTrueF(ok, "only string schema as map key for type %s but got %s", t, key)
		}
		return jsonschema.RecordOf(key, SchemaFromType(ctx, t.Elem(), opt.WithDecl(false)))
	case reflect.Struct:
		ss := jsonschema.ObjectOf(nil)

		fs, err := scanner.Structs.Scan(t)
		must.NoErrorF(err, "failed to extract schema for %s", t)

		for f := range fs.Range {
			propSchema := toPropSchema(ctx, f, opt)
			if propSchema != nil {
				ss.SetProperty(f.Name, propSchema, !(f.Omitempty || f.Omitzero))
			}
		}

		v := reflect.New(t).Interface()
		if manifest, ok := v.(K8sObjectKindGetter); ok {
			apiVersion, kind := manifest.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

			if kind != "" {
				if _, ok := ss.Properties.Get("kind"); ok {
					ss.SetProperty("kind", &jsonschema.EnumType{
						Enum: []any{kind},
					}, false)
				}
			}

			if apiVersion != "" {
				if _, ok := ss.Properties.Get("apiVersion"); ok {
					ss.SetProperty("apiVersion", &jsonschema.EnumType{
						Enum: []any{apiVersion},
					}, false)
				}
			}
		}

		if manifest, ok := v.(K8sKindGetter); ok {
			if kind := manifest.GetKind(); kind != "" {
				if _, ok := ss.Properties.Get("kind"); ok {
					ss.SetProperty("kind", &jsonschema.EnumType{Enum: []any{kind}}, false)
				}
			}
		}

		if manifest, ok := v.(K8sAPIVersionGetter); ok {
			if apiVersion := manifest.GetAPIVersion(); apiVersion != "" {
				if _, ok := ss.Properties.Get("apiVersion"); ok {
					ss.SetProperty("apiVersion", &jsonschema.EnumType{Enum: []any{apiVersion}}, false)
				}
			}
		}
		return ss
	default:
		panic(fmt.Errorf("unsupported schema type for: %s", t))
	}
}
