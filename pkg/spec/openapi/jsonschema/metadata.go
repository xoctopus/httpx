package jsonschema

import "maps"

const (
	XEnumLabels    = `x-enum-labels`
	XGoType        = `x-go-type`
	XGoVendorType  = `x-go-vendor-type`
	XGoStarLevel   = `x-go-star-level`
	XGoFieldName   = `x-go-field-name`
	XTagValidate   = `x-tag-validate`
	XPatternErrMsg = `x-pattern-err-msg`
)

type Metadata struct {
	Title       string `json:"title,omitzero"`
	Description string `json:"description,omitzero"`
	Default     any    `json:"default,omitzero"`
	WriteOnly   *bool  `json:"writeOnly,omitzero"`
	ReadOnly    *bool  `json:"readOnly,omitzero"`
	Examples    []any  `json:"examples,omitzero"`
	Deprecated  *bool  `json:"deprecated,omitzero"`

	Ext
}

func (v *Metadata) GetMetadata() *Metadata {
	return v
}

func (v *Metadata) DeepCopy() *Metadata {
	if v == nil {
		return nil
	}
	out := new(Metadata)
	v.DeepCopyInto(out)
	return out
}

func (v *Metadata) DeepCopyInto(out *Metadata) {
	out.Title = v.Title
	out.Description = v.Description
	out.Default = v.Default
	out.WriteOnly = v.WriteOnly
	out.ReadOnly = v.ReadOnly
	if v.Examples != nil {
		i, o := &v.Examples, &out.Examples
		*o = make([]any, len(*i))
		copy(*o, *i)
	}
	out.Deprecated = v.Deprecated
	v.Ext.DeepCopyInto(&out.Ext)
}

type Ext struct {
	Extensions map[string]any `json:",inline"`
}

func (v *Ext) DeepCopy() *Ext {
	if v == nil {
		return nil
	}
	out := new(Ext)
	v.DeepCopyInto(out)
	return out
}

func (v *Ext) DeepCopyInto(out *Ext) {
	if i := v.Extensions; i != nil {
		o := make(map[string]any, len(i))
		maps.Copy(o, out.Extensions)
		maps.Copy(o, i)
		out.Extensions = o
	}
}

func (v Ext) Merge(m Ext) Ext {
	ext := Ext{}

	for k := range v.Extensions {
		ext.AddExtension(k, v.Extensions[k])
	}

	for k := range m.Extensions {
		ext.AddExtension(k, v.Extensions[k])
	}

	return ext
}

func (v *Ext) AddExtension(key string, value any) {
	if value == nil {
		return
	}
	if v.Extensions == nil {
		v.Extensions = make(map[string]any)
	}
	v.Extensions[key] = value
}

func (v *Ext) GetExtension(key string) (any, bool) {
	if v.Extensions == nil {
		return nil, false
	}

	e, ok := v.Extensions[key]
	return e, ok
}
