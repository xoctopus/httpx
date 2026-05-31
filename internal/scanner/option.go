package scanner

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"

	"github.com/xoctopus/x/reflectx"
)

func ParseFieldOptions(sf reflect.StructField) (FieldOptions, bool, error) {
	return parseFieldOptions(sf, "")
}

func parseFieldOptions(sf reflect.StructField, prefix string) (FieldOptions, bool, error) {
	if v, ok := sf.Tag.Lookup("in"); ok && v == "body" {
		if _, exists := sf.Tag.Lookup("name"); !exists {
			sf.Tag += reflect.StructTag(fmt.Sprintf(` name:%q`, v))
		}
	}

	tag := reflectx.ParseTag(
		sf.Tag,
		reflectx.WithOptionSplitter(':'),
		reflectx.WithExpectFlags("json", "name"),
	)

	flag := cmp.Or(tag.Get("name"), tag.Get("json"))
	var (
		options *FieldOptions
		ignored bool
		format  string
		casing  = Casing(0)
	)

	if flag == nil || flag.Name() == "-" {
		options = &FieldOptions{Name: sf.Name}
		ignored = flag != nil && flag.Name() == "-"
	} else {
		if flag.HasOption("nocase") {
			casing |= IgnoreCase
		}
		if flag.HasOption("strictcase") {
			casing |= StrictCase
		}
		if casing == IgnoreCase|StrictCase {
			return FieldOptions{}, false, errors.New("couldn't have both `nocase` and `strictcase`")
		}

		if o := flag.Option("format"); o != nil {
			format = o.Unquoted()
		}

		name := cmp.Or(flag.Name(), sf.Name)
		if len(prefix) > 0 {
			name = prefix + "." + name
		}

		options = &FieldOptions{
			Name:       name,
			HasName:    len(flag.Name()) > 0,
			Casing:     casing,
			Inline:     flag.HasOption("inline"),
			Unknown:    flag.HasOption("unknown"),
			Embedded:   flag.HasOption("embedded"),
			Omitzero:   flag.HasOption("omitzero"),
			Omitempty:  flag.HasOption("omitempty"),
			String:     flag.HasOption("string"),
			Format:     format,
			StringItem: false,
		}
	}
	if options.Embedded && (options.Inline || options.Unknown) {
		return FieldOptions{}, false, errors.New("couldn't have both `inline` and `embedded`")
	}
	if !options.String {
		options.String = CanUnmarshalByString(sf.Type)
	}
	if !options.String && sf.Type.Kind() == reflect.Slice {
		options.StringItem = CanUnmarshalByString(sf.Type.Elem())
	}
	return *options, ignored, nil
}

// FieldOptions see encoding/json/v2.fieldOptions.
// nameNeedEscape is unsupported
type FieldOptions struct {
	Name       string
	HasName    bool
	Casing     Casing
	Inline     bool
	Unknown    bool
	Omitzero   bool
	Omitempty  bool
	String     bool
	Format     string
	StringItem bool

	// Embedded not in v2.fieldOptions. for parsing duplicated query kvs
	// eg: createdAt.gt=x&createdAt.lte=y&updatedAt.gt=x&updatedAt.lte=y
	// the gt, lte in createdAt and updatedAt is same type but prefix.
	// This is a temporary plan
	Embedded bool
}

func (o *FieldOptions) StrictCase() bool {
	return o.Casing == StrictCase
}

type Casing uint8

const (
	IgnoreCase Casing = 1
	StrictCase Casing = 2
)
