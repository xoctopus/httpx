package extractors

import (
	"context"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/xoctopus/x/docx"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/slicex"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
	"github.com/xoctopus/httpx/pkg/spec/openapi/jsonschema"
)

func SetTitleOrDescription(metadata *jsonschema.Metadata, lines []string) {
	if metadata == nil {
		return
	}

	if len(lines) > 0 {
		metadata.Title = strings.TrimSpace(lines[0])

		if len(lines) > 1 {
			lines = slicex.FilterMapping(lines, func(from string) (string, bool) {
				if strings.HasPrefix(from, "openapi:") {
					return "", false
				}
				return from, true
			})
			metadata.Description = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}
}

type FieldExclude func(fields ...string)

type FieldFilter struct {
	Exclude []string
	Include []string
}

var fieldFilters sync.Map

func RegisterFieldFilter(t reflect.Type, fieldFilter FieldFilter) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	fieldFilters.Store(t, fieldFilter)
}

func FieldShouldPick(t reflect.Type, fieldName string) bool {
	if filter, ok := fieldFilters.Load(t); ok {
		ff := filter.(FieldFilter)

		if len(ff.Include) > 0 {
			return slices.Contains(ff.Include, fieldName)
		}

		if len(ff.Exclude) > 0 {
			return !slices.Contains(ff.Exclude, fieldName)
		}

		return false
	}

	return true
}

func toPropSchema(ctx context.Context, sf *scanner.Field, opt Opt) jsonschema.Schema {
	if !FieldShouldPick(sf.Type, sf.FieldName) {
		return nil
	}

	var docs []string

	if opt.Doc != nil {
		for _, name := range []string{
			sf.FieldName,
			sf.Name,
		} {
			if fieldDesc := opt.Doc[name]; fieldDesc != "" {
				stringEnum := pickStringEnumFromDesc(fieldDesc)
				if len(stringEnum) > 0 {
					opt = opt.WithEnumInDoc(stringEnum)
				}

				if i := strings.Index(fieldDesc, "."); i > 0 {
					docs = []string{fieldDesc[0:i], fieldDesc[i+1:]}
				} else if i := strings.Index(fieldDesc, "\n"); i > 0 {
					docs = []string{fieldDesc[0:i], fieldDesc[i+1:]}
				} else {
					docs = []string{fieldDesc, ""}
				}
			}
		}
	}

	propSchema := SchemaFromType(ctx, sf.Type, opt.WithDecl(false))
	if propSchema != nil {
		option := validation.Option{Type: sf.Type}
		text := sf.Tag.Get("validate")
		if len(text) > 0 {
			option.Rule = rule.MustCompile(text)
		}
		s, err := PatchSchemaValidation(propSchema, option)
		must.NoErrorF(err, "failed to patch validation for %s.%s from %s", sf.Type, sf.FieldName, text)

		SetTitleOrDescription(s.GetMetadata(), docs)
		if dp, ok := docx.ProviderFrom(ctx); ok {
			if lines, ok := dp.DocOf(sf.FieldName); ok {
				SetTitleOrDescription(s.GetMetadata(), lines)
			}
		}
		s.GetMetadata().AddExtension(jsonschema.XGoFieldName, sf.FieldName)
		return s
	}

	return nil
}

func pickStringEnumFromDesc(d string) []string {
	parts := strings.SplitSeq(d, ".")
	for p := range parts {
		line := strings.TrimSpace(p)
		if strings.HasPrefix(line, "One of") {
			enumValues := strings.Split(line[len("One of")+1:], ",")
			for i := range enumValues {
				enumValues[i] = strings.TrimSpace(enumValues[i])
			}
			return enumValues
		}
		if strings.HasPrefix(line, "Can be") {
			enumValues := strings.Split(line[len("Can be")+1:], " or ")
			for i := range enumValues {
				enumValues[i] = strings.TrimSpace(enumValues[i])
				if len(enumValues[i]) > 0 {
					if enumValues[i][0] == '"' {
						enumValues[i], _ = strconv.Unquote(enumValues[i])
					}
				}
			}
			return enumValues
		}
	}

	return nil
}

type (
	K8sObjectKindGetter interface {
		GetObjectKind() schema.ObjectKind
	}
	K8sKindGetter interface {
		GetKind() string
	}
	K8sAPIVersionGetter interface {
		GetAPIVersion() string
	}
)

type TypeName string

func (t TypeName) RefString() string {
	return string(t)
}

type Opt struct {
	Decl      bool
	Doc       map[string]string
	EnumInDoc []string
}

func (o Opt) WithDecl(decl bool) Opt {
	o.Decl = decl
	return o
}

func (o Opt) WithDoc(doc map[string]string) Opt {
	o.Doc = doc
	return o
}

func (o Opt) WithEnumInDoc(enumInDoc []string) Opt {
	o.EnumInDoc = enumInDoc
	return o
}
