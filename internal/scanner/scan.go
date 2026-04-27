package scanner

import (
	"cmp"
	"iter"
	"reflect"
	"slices"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/syncx"
)

var Structs = &structs{
	fields: syncx.NewXmap[reflect.Type, *Struct](),
}

type structs struct {
	fields syncx.Map[reflect.Type, *Struct]
}

func (v *structs) Scan(typ reflect.Type) (*Struct, error) {
	must.BeTrueF(
		typ.Kind() == reflect.Struct,
		"invalid input type: %s. expect Struct", typ,
	)

	if fs, ok := v.fields.Load(typ); ok {
		return fs, nil
	}

	fs, err := scan(typ, "")
	if err != nil {
		return nil, err
	}
	v.fields.Store(typ, fs)

	return fs, nil
}

func scan(root reflect.Type, prefix string) (*Struct, error) {
	type entry struct {
		typ reflect.Type
		idx []int
		sub bool
	}
	var (
		entered = []entry{{root, nil, true}}
		seen    = map[reflect.Type]bool{root: true}
		index   = 0
		fields  = make([]*Field, 0)
		inlined = make([]*Field, 0)
	)

	for index < len(entered) {
		e := entered[index]
		index++

		t := e.typ
		for i := range t.NumField() {
			f := t.Field(i)
			options, ignored, err := parseFieldOptions(f, prefix)
			if err != nil {
				return nil, err
			}
			if ignored {
				continue
			}

			fi := &Field{
				FieldOptions: options,
				FieldName:    f.Name,
				Type:         f.Type,
				Tag:          f.Tag,
				index:        make([]int, 0, len(e.idx)+1),
			}
			fi.index = append(fi.index, e.idx...)
			fi.index = append(fi.index, i)

			if f.Anonymous && !fi.HasName {
				fi.Inline = true
			}

			// The embedded option is only permitted within query tags and must
			// be explicitly specified. see Testdata.InQuery
			if fi.Embedded && fi.Tag.Get("in") == "query" {
				name := options.Name
				if len(prefix) > 0 {
					name = prefix + "." + name
				}
				se, erre := scan(fi.Type, name)
				if erre != nil {
					return nil, erre
				}
				for fe := range se.Range {
					fields = append(fields, fe)
				}
				continue
			}

			if fi.Inline || fi.Unknown || fi.Embedded {
				tf := f.Type
				for tf.Kind() == reflect.Pointer && tf.Name() == "" {
					tf = tf.Elem()
				}
				if tf.Kind() == reflect.Struct {
					if fi.Unknown {
						continue
					}
					if e.sub {
						entered = append(entered, entry{
							typ: tf,
							idx: fi.index,
							sub: !seen[tf],
						})
					}
					seen[tf] = true
					continue
				}
				inlined = append(inlined, fi)
			} else {
				fields = append(fields, fi)
			}
		}
	}
	fs := &Struct{
		flattened: slices.SortedFunc(slices.Values(fields), func(a, b *Field) int {
			return cmp.Compare(a.index[0], b.index[0])
		}),
		names:   make(map[string]*Field, len(fields)),
		locates: make(map[locate]*Field, len(fields)),
		located: make(map[string][]*Field, len(fields)),
	}

	for _, f := range fs.flattened {
		// TODO what if parsed field name(or key) conflict
		// ref: https://github.com/go-json-experiment/json/issues/189
		// eg:
		//	type A struct {
		//		V1 int `json:"value"
		//	}
		//	type B struct {
		//		A
		//		V2 int `json:"value"
		//	}
		fs.names[f.Name] = f

		if loc := f.Tag.Get("in"); loc != "" {
			fs.locates[locate{loc: loc, name: f.Name}] = f
			fs.located[loc] = append(fs.located[loc], f)
		}
	}

	if n := len(inlined); n == 1 || (n > 1 && len(inlined[0].index) != len(inlined[1].index)) {
		fs.inlined = inlined[0]
	}

	return fs, nil
}

type Field struct {
	FieldOptions

	FieldName string
	Tag       reflect.StructTag
	Type      reflect.Type

	index []int
}

func (f *Field) Multiple() bool {
	kind := f.Type.Kind()
	return !f.String && (kind == reflect.Slice || kind == reflect.Array)
}

func (f *Field) GetOrNewAt(v reflect.Value) reflect.Value {
	if len(f.index) == 1 {
		return v.Field(f.index[0])
	}

	for i, idx := range f.index {
		if i > 0 {
			if v.Kind() == reflect.Pointer {
				if v.IsNil() {
					v.Set(reflect.New(v.Type().Elem()))
				}
				v = v.Elem()
			}
		}
		v = v.Field(idx)
	}

	return v
}

type Struct struct {
	flattened []*Field
	names     map[string]*Field
	inlined   *Field
	located   map[string][]*Field
	locates   map[locate]*Field
}

func (fs *Struct) Len() int {
	return len(fs.flattened)
}

func (fs *Struct) Lookup(loc, name string) (*Field, bool) {
	f, ok := fs.locates[locate{loc, name}]
	return f, ok
}

func (fs *Struct) FieldsIn(loc string) iter.Seq[*Field] {
	return func(yield func(*Field) bool) {
		for _, sf := range fs.located[loc] {
			if !yield(sf) {
				return
			}
		}
	}
}

func (fs *Struct) FieldsInCookie() iter.Seq[*Field] {
	return fs.FieldsIn("cookie")
}

func (fs *Struct) FieldsInHeader() iter.Seq[*Field] {
	return fs.FieldsIn("header")
}

func (fs *Struct) FieldsInPath() iter.Seq[*Field] {
	return fs.FieldsIn("path")
}

func (fs *Struct) FieldsInQuery() iter.Seq[*Field] {
	return fs.FieldsIn("query")
}

func (fs *Struct) Range(fn func(f *Field) bool) {
	for _, f := range fs.flattened {
		if !fn(f) {
			break
		}
	}
}

func (fs *Struct) RangeIn(loc string) iter.Seq[*Field] {
	return func(yield func(*Field) bool) {
		for _, sf := range fs.located[loc] {
			if !yield(sf) {
				return
			}
		}
	}
}

func (fs *Struct) Inlined() (*Field, bool) {
	return fs.inlined, fs.inlined != nil
}

type locate struct {
	loc  string
	name string
}
