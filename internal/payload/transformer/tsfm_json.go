package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"reflect"

	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/pkg/validation"
)

func init() {
	Register(&_jsonP{})
}

type _jsonP struct{}

func (p *_jsonP) Names() []string {
	return []string{"application/json", "json"}
}

func (p *_jsonP) Transformer() (Transformer, error) {
	return &jsonT{media: p.Names()[0]}, nil
}

type jsonT struct {
	media string
}

func (p *jsonT) Media() string {
	return p.media
}

func (p *jsonT) Into(ctx context.Context, r io.ReadCloser, v any) error {
	defer func() { _ = r.Close() }()

	u := scanner.UnwrapField(v)
	rv, ok := u.(reflect.Value)
	if ok {
		u = rv.Interface()
	}

	if x, ok := u.(json.Unmarshaler); ok {
		raw, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		return x.UnmarshalJSON(raw)
	}

	return validation.UnmarshalReader(r, u)
}

func (p *jsonT) Prepare(ctx context.Context, v any) (content.Content, error) {
	c := content.NewBuilderFrom(p.media, map[string]string{"charset": "utf-8"})

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	if x, ok := v.(json.Marshaler); ok {
		// avoid trim \n
		raw, err := x.MarshalJSON()
		if err != nil {
			return nil, err
		}

		c.SetContentLength(int64(len(raw)))
		c.SetReadCloser(io.NopCloser(bytes.NewBuffer(raw)))

		return c, nil
	}

	c.SetReadCloser(content.AsReadCloser(ctx, func(w io.WriteCloser) func() error {
		return func() error {
			defer func() { _ = w.Close() }()
			return validation.MarshalWrite(w, v)
		}
	}))

	return c, nil
}
