package transformer

import (
	"bytes"
	"context"
	"io"
	"mime"
	"reflect"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/pkg/validation"
)

func init() {
	Register(&octetP{})
}

type octetP struct{}

func (p *octetP) Names() []string {
	return []string{
		"application/octet-stream",
		"octet-stream",
		"octet",
	}
}

func (p *octetP) Transformer() (Transformer, error) {
	return &octetT{
		media: p.Names()[0],
	}, nil
}

type octetT struct {
	media string
}

func (p *octetT) Media() string {
	return p.media
}

func (p *octetT) Into(_ context.Context, r io.ReadCloser, v any) error {
	u := scanner.UnwrapField(v)

	rv, ok := u.(reflect.Value)
	if ok {
		u = rv.Interface()
	}

	if x, ok := u.(*io.ReadCloser); ok {
		*x = r
		return nil
	}

	header, ok := r.(content.WithHeader)
	if ok {
		if x, ok := u.(content.FilenameModifier); ok {
			_, params, err := mime.ParseMediaType(header.Header().Get("Content-Disposition"))
			if err == nil {
				x.SetFilename(params["filename"])
			}
		}

		if x, ok := u.(content.MediaTypeModifier); ok {
			x.SetContentType(header.Header().Get("Content-Type"))
		}
	}

	if x, ok := u.(content.ReaderFrom); ok {
		_, err := x.ReadFrom(r)
		return err
	}

	defer func() { _ = r.Close() }()

	var err error
	switch x := u.(type) {
	case io.ReaderFrom:
		_, err = x.ReadFrom(r)
	case io.Writer:
		_, err = io.Copy(x, r)
	case *[]byte:
		*x, err = io.ReadAll(r)
	default:
		var data []byte
		if data, err = io.ReadAll(r); err != nil {
			return err
		}
		if data, err = jsontext.AppendQuote(nil, data); err != nil {
			return err
		}
		err = json.Unmarshal(data, u)
	}
	return err
}

func (p *octetT) Prepare(_ context.Context, v any) (content.Content, error) {
	b := content.NewBuilder(p.media)

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	if v == nil {
		b.SetContentLength(0)
		b.SetReadCloser(io.NopCloser(bytes.NewBuffer(nil)))
		return b, nil
	}

	switch x := v.(type) {
	case io.ReadCloser:
		b.SetReadCloser(x)
	case io.Reader:
		b.SetReadCloser(io.NopCloser(x))
	case []byte:
		b.SetContentLength(int64(len(x)))
		b.SetReadCloser(io.NopCloser(bytes.NewBuffer(x)))
	case string:
		b.SetContentLength(int64(len(x)))
		b.SetReadCloser(io.NopCloser(bytes.NewBufferString(x)))
	default:
		data, err := validation.Marshal(x)
		if err != nil {
			return nil, err
		}
		data, err = jsontext.AppendUnquote(nil, data)
		if err != nil {
			return nil, err
		}
		b.SetContentLength(int64(len(data)))
		b.SetReadCloser(io.NopCloser(bytes.NewBuffer(data)))
	}
	return b, nil
}
