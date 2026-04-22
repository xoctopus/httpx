package scanner

type WrappedField interface {
	Unwrap() any
	Field() *Field
}

func WrapField(v any, f *Field) any {
	return &wrappedField{
		v: v,
		f: f,
	}
}

func UnwrapField(v any) any {
	if x, ok := v.(WrappedField); ok {
		return x.Unwrap()
	}
	return v
}

type wrappedField struct {
	v any
	f *Field
}

func (wf *wrappedField) Unwrap() any {
	return wf.v
}

func (wf *wrappedField) Field() *Field {
	return wf.f
}
