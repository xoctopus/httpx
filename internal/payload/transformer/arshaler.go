package transformer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"sync"

	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/payload"
	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/payload/path"
	"github.com/xoctopus/httpx/internal/request"
	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/internal/validation"
)

type Request = request.Request

type ValueIterator interface {
	Values(*scanner.Field) iter.Seq[reflect.Value]
}

type MutableValueIterator interface {
	ValuesForModification(*scanner.Field, int) iter.Seq2[int, reflect.Value]
}

type ValueMarshaler interface {
	MarshalValues(ctx context.Context, f *scanner.Field) (values []string, err error)
}

type ValueUnmarshaler interface {
	UnmarshalValues(ctx context.Context, f *scanner.Field, values []string) error
}

type ReaderUnmarshaler interface {
	UnmarshalReaders(ctx context.Context, f *scanner.Field, readers []io.ReadCloser) error
}

type RequestMarshaler interface {
	MarshalRequest(ctx context.Context, method string, segments path.Segments) (*http.Request, error)
}

type RequestUnmarshaler interface {
	UnmarshalUnderlying(*http.Request) error
	UnmarshalRequest(r Request) error
}

type parameter struct {
	reflect.Value
}

func (p *parameter) Values(f *scanner.Field) iter.Seq[reflect.Value] {
	if !f.String {
		if f.Type.Kind() == reflect.Slice || f.Type.Kind() == reflect.Array {
			rv := f.GetOrNewAt(p.Value)
			return func(yield func(v reflect.Value) bool) {
				for i := 0; i < rv.Len(); i++ {
					if !yield(rv.Index(i)) {
						return
					}
				}
			}
		}
	}

	return func(yield func(v reflect.Value) bool) {
		if !yield(f.GetOrNewAt(p.Value)) {
			return
		}
	}
}

func (p *parameter) ValuesForModification(f *scanner.Field, n int) iter.Seq2[int, reflect.Value] {
	if f.Multiple() {
		if n == 0 {
			return func(yield func(int, reflect.Value) bool) {}
		}

		rv := f.GetOrNewAt(p.Value)
		if rv.Cap() < n {
			rv.Grow(n)
			rv.SetLen(n)
		}

		return func(yield func(i int, v reflect.Value) bool) {
			for i := 0; i < rv.Len(); i++ {
				if !yield(i, rv.Index(i).Addr()) {
					return
				}
			}
		}
	}

	return func(yield func(i int, v reflect.Value) bool) {
		if !yield(0, f.GetOrNewAt(p.Value).Addr()) {
			return
		}
	}
}

func (p *parameter) UnmarshalValues(ctx context.Context, f *scanner.Field, values []string) error {
	if len(values) == 0 {
		va, err := validation.NewFromStructField(f)
		if err != nil {
			return err
		}

		if x, ok := va.(validation.WithDefaults); ok {
			if v := x.Defaults(); len(v) > 0 {
				d := jsontext.Value(v)
				if d.Kind() == jsontext.STRING {
					unquoted, err := jsontext.AppendUnquote(nil, d)
					if err != nil {
						return err
					}
					d = unquoted
				}
				values = []string{string(d)}
			}
		}
	}

	readers := make([]io.ReadCloser, len(values))
	for i := range values {
		readers[i] = io.NopCloser(bytes.NewBufferString(values[i]))
	}
	return p.UnmarshalReaders(ctx, f, readers)
}

func (p *parameter) UnmarshalReaders(ctx context.Context, f *scanner.Field, readers []io.ReadCloser) error {
	for i, rv := range p.ValuesForModification(f, len(readers)) {
		t, err := New(rv.Elem().Type(), f.Tag.Get("mime"), ForUnmarshalling)
		if err != nil {
			return err
		}

		if i < len(readers) {
			err = t.Into(ctx, readers[i], scanner.WrapField(rv, f))
		} else {
			if !(f.Omitempty || f.Omitzero) {
				err = codex.New(validation.ERROR__MISSING_REQUIRED)
			}
		}

		if err == nil {
			continue
		}

		if f.Multiple() {
			return validation.WrapPosition(err, fmt.Sprintf("%s.%d", f.FieldName, i))
		}
		return validation.WrapPosition(err, f.FieldName)
	}
	return nil
}

func (p *parameter) MarshalValues(ctx context.Context, f *scanner.Field) (values []string, err error) {
	for rv := range p.Values(f) {
		if rv.IsZero() {
			if f.Omitempty || f.Omitzero {
				continue
			}
		}

		var t Transformer

		t, err = New(rv.Type(), f.Tag.Get("mime"), ForMarshalling)
		if err != nil {
			return nil, err
		}

		b := bytes.NewBuffer(nil)

		if err = func() error {
			c, e := t.Prepare(ctx, rv)
			if e != nil {
				return e
			}
			defer func() { _ = c.Close() }()
			_, e = io.Copy(b, c)
			return e
		}(); err != nil {
			return nil, err
		}

		values = append(values, b.String())
	}

	return values, err
}

func (p *parameter) MarshalRequest(ctx context.Context, method string, segments path.Segments) (*http.Request, error) {
	s, err := scanner.Structs.Scan(p.Type())
	if err != nil {
		return nil, err
	}

	var (
		body    content.Content
		query   = url.Values{}
		header  = http.Header{}
		params  = path.Values{}
		cookies = url.Values{}
	)

	for f := range s.RangeIn("body") {
		rv := f.GetOrNewAt(p.Value)

		c, err := New(f.Type, f.Tag.Get("mime"), ForMarshalling)
		if err != nil {
			return nil, err
		}
		body, err = c.Prepare(ctx, rv)
		if err != nil {
			return nil, err
		}

		// skip following for only one body
		break
	}

	for f := range s.RangeIn("path") {
		values, err := p.MarshalValues(ctx, f)
		if err != nil {
			return nil, err
		}

		if len(values) > 0 {
			params[f.Name] = values[0]
		}
	}

	for f := range s.RangeIn("query") {
		values, err := p.MarshalValues(ctx, f)
		if err != nil {
			return nil, err
		}

		if values != nil {
			query[f.Name] = values
		}
	}

	for f := range s.RangeIn("header") {
		values, err := p.MarshalValues(ctx, f)
		if err != nil {
			return nil, err
		}

		if values != nil {
			header[textproto.CanonicalMIMEHeaderKey(f.Name)] = values
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, segments.Encode(params), body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		if t := body.Type(); t != "" {
			header.Set("Content-Type", t)
		}

		if l := body.Length(); l > -1 {
			req.ContentLength = l
			header.Set("Content-Length", strconv.FormatInt(l, 10))
		}
	}
	req.Header = header
	if len(query) > 0 {
		req.URL.RawQuery = query.Encode()
	}

	for sf := range s.RangeIn("cookie") {
		values, err := p.MarshalValues(ctx, sf)
		if err != nil {
			return nil, err
		}

		if len(values) > 0 {
			cookies[sf.Name] = values
		}
	}

	if n := len(cookies); n > 0 {
		names := make([]string, n)
		i := 0
		for name := range cookies {
			names[i] = name
			i++
		}
		sort.Strings(names)

		for _, name := range names {
			values := cookies[name]

			for i := range values {
				req.AddCookie(&http.Cookie{
					Name:  name,
					Value: values[i],
				})
			}
		}
	}

	return req, err
}

func (p *parameter) UnmarshalUnderlying(r *http.Request) error {
	return p.UnmarshalRequest(request.From(r))
}

func (p *parameter) UnmarshalRequest(r request.Request) error {
	s, err := scanner.Structs.Scan(p.Type())
	if err != nil {
		return err
	}

	once := &sync.Once{}

	for loc := range payload.Locations {
		for f := range s.RangeIn(loc) {
			if loc == "body" {
				once.Do(func() {
					body := r.Body()
					if body == nil {
						return
					}
					var (
						t  Transformer
						rv = f.GetOrNewAt(p.Value)
					)
					t, err = New(f.Type, f.Tag.Get("mime"), ForUnmarshalling)
					if err != nil {
						return
					}
					err = t.Into(r.Context(), body, rv.Addr())
				})
			} else {
				err = p.UnmarshalValues(r.Context(), f, r.ValuesIn(loc, f.Name))
			}

			if err != nil {
				return validation.WrapLocationPosition(err, loc, f.Name)
			}
		}
	}

	return nil
}
