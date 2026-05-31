package transformer

import (
	"context"
	"io"
	"net/url"
	"reflect"

	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/pkg/validation"
)

func init() {
	Register(&urlencodedP{})
}

type urlencodedP struct{}

func (p *urlencodedP) Names() []string {
	return []string{
		"application/x-www-form-urlencoded",
		"form",
		"urlencoded",
		"url-encoded",
	}
}

func (p *urlencodedP) Transformer() (Transformer, error) {
	return &urlencodedT{
		media: p.Names()[0],
	}, nil
}

type urlencodedT struct {
	media string
}

func (p *urlencodedT) Media() string {
	return p.media
}

func (p *urlencodedT) Into(_ context.Context, r io.ReadCloser, v any) error {
	defer func() { _ = r.Close() }()

	return content.Pipe(
		func(r io.Reader) error {
			return validation.UnmarshalReader(r, v)
		},
		func(w io.Writer) error {
			data, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			values, err := url.ParseQuery(string(data))
			if err != nil {
				return err
			}
			return validation.MarshalWrite(w, values)
		},
	)
}

func (p *urlencodedT) Prepare(ctx context.Context, v any) (content.Content, error) {
	b := content.NewBuilderFrom(p.media, map[string]string{"param": "value"})

	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	rc := content.AsReadCloser(
		ctx,
		func(wc io.WriteCloser) func() error {
			uw := &urlencodedWriter{Writer: wc}

			return func() error {
				defer func() { _ = wc.Close() }()

				return content.Pipe(
					func(r io.Reader) error {
						return uw.WriteDecoder(jsontext.NewDecoder(r))
					},
					func(w io.Writer) error {
						return validation.MarshalWrite(w, v)
					},
				)
			}
		},
	)
	b.SetReadCloser(rc)

	return b, nil
}

type urlencodedWriter struct {
	io.Writer
}

func (w *urlencodedWriter) WriteDecoder(d *jsontext.Decoder) error {
	tok, err := d.ReadToken()
	if err != nil {
		return err
	}
	if tok.Kind() != jsontext.OBJECT {
		return nil
	}

	write := func(i int, key string, val jsontext.Value) (written bool, err error) {
		if i > 0 {
			if _, err = w.Write([]byte("&")); err != nil {
				return false, err
			}
		}

		switch val.Kind() {
		case jsontext.NULL:
			return false, nil
		case jsontext.FALSE, jsontext.TRUE, jsontext.NUMBER:
			goto FINISH
		case jsontext.STRING:
			val, err = jsontext.AppendUnquote(nil, val)
			if err != nil {
				return false, err
			}
			goto FINISH
		default:
			return false, nil
		}
	FINISH:
		if _, err = w.Write([]byte(url.QueryEscape(key) + "=")); err != nil {
			return false, err
		}
		if _, err = w.Write([]byte(url.QueryEscape(string(val)))); err != nil {
			return false, err
		}
		return true, nil
	}

	i, written := 0, false
	for d.PeekKind() != jsontext.OBJECT_E {
		tok, err = d.ReadToken()
		if err != nil {
			return err
		}
		key := tok.String()

		var val jsontext.Value
		switch kind := d.PeekKind(); kind {
		case jsontext.ARRAY:
			if _, err = d.ReadToken(); err != nil {
				return err
			}
			for d.PeekKind() != jsontext.ARRAY_E {
				val, err = d.ReadValue()
				if err != nil {
					return err
				}
				written, err = write(i, key, val)
				if err != nil {
					return err
				}
				if written {
					i++
				}
			}
			if _, err = d.ReadToken(); err != nil {
				return err
			}
		default:
			val, err = d.ReadValue()
			if err != nil {
				return err
			}
			written, err = write(i, key, val)
			if err != nil {
				return err
			}
			if written {
				i++
			}
		}
	}
	_, err = d.ReadToken()

	return err
}
