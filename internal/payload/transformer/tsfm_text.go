package transformer

import (
	"bytes"
	"context"
	"io"
	"reflect"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/pkg/validation"
)

func init() {
	Register(&textP{})
}

type textP struct{}

func (p *textP) Names() []string {
	return []string{
		"text/plain",
		"plain",
		"text",
		"txt",
	}
}

func (p *textP) Transformer() (Transformer, error) {
	return &textT{
		media: p.Names()[0],
	}, nil
}

type textT struct {
	media string
}

func (p *textT) Media() string {
	return p.media
}

func (p *textT) Into(_ context.Context, r io.ReadCloser, v any) error {
	defer func() { _ = r.Close() }()

	u := scanner.UnwrapField(v)
	rv, ok := u.(reflect.Value)
	if ok {
		u = rv.Interface()
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	switch x := u.(type) {
	case *[]byte:
		*x = data
	default:
		data, err = jsontext.AppendQuote(nil, data)
		if err != nil {
			return err
		}
		err = validation.Unmarshal(data, x)
	}
	return err
}

func (p *textT) Prepare(_ context.Context, v any) (content.Content, error) {
	b := content.NewBuilderFrom(p.media, map[string]string{"charset": "utf-8"})

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	switch x := v.(type) {
	case []byte:
		b.SetContentLength(int64(len(x)))
		b.SetReadCloser(io.NopCloser(bytes.NewBuffer(x)))
		return b, nil
	case string:
		b.SetContentLength(int64(len(x)))
		b.SetReadCloser(io.NopCloser(bytes.NewBufferString(x)))
		return b, nil
	default:
		data, err := validation.Marshal(v)
		if err != nil {
			return nil, err
		}
		data, err = jsontext.AppendUnquote(nil, data)
		if err != nil {
			return nil, err
		}
		b.SetContentLength(int64(len(data)))
		b.SetReadCloser(io.NopCloser(bytes.NewBuffer(data)))
		return b, nil
	}
}
