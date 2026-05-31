package transformer

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strconv"

	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/request"
	"github.com/xoctopus/httpx/internal/scanner"
)

func init() {
	Register(&multipartP{})
}

type multipartP struct{}

func (p *multipartP) Names() []string {
	return []string{
		"multipart/form-data",
		"multipart",
		"form-data",
	}
}

func (p *multipartP) Transformer() (Transformer, error) {
	return &multipartT{
		media: p.Names()[0],
	}, nil
}

type multipartT struct {
	media string
}

func (t *multipartT) Media() string {
	return t.media
}

func (t *multipartT) Into(ctx context.Context, r io.ReadCloser, v any) error {
	defer func() { _ = r.Close() }()

	f := scanner.UnwrapField(v)

	h := http.Header{}
	if x, ok := r.(content.WithHeader); ok {
		h = x.Header()
	}

	_, params, err := mime.ParseMediaType(h.Get("Content-Type"))
	if err != nil {
		return err
	}

	reader := multipart.NewReader(r, params["boundary"])
	form, err := reader.ReadForm(content.MaxBufferSize)
	if err != nil {
		return err
	}

	rv, ok := f.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(f)
	}

	if rv.Kind() != reflect.Pointer {
		return errors.New("target must be pointer value")
	}

	pv := NewArshaler(rv.Elem())

	s, err := scanner.Structs.Scan(pv.Type())
	if err != nil {
		return err
	}

	for sf := range s.Range {
		if sf.Type.Implements(content.TWithFilename) ||
			sf.Multiple() && sf.Type.Elem().Implements(content.TWithFilename) {

			headers := form.File[sf.Name]
			if len(headers) == 0 {
				continue
			}
			readers := make([]io.ReadCloser, len(headers))
			for i, fh := range headers {
				file, err := fh.Open()
				if err != nil {
					return err
				}
				readers[i] = request.ReadCloserWithHeader(file, http.Header(fh.Header))
			}

			if err = pv.UnmarshalReaders(ctx, sf, readers); err != nil {
				return err
			}
			continue
		}

		if err = pv.UnmarshalValues(ctx, sf, form.Value[sf.Name]); err != nil {
			return err
		}
	}

	return nil
}

func (t *multipartT) Prepare(ctx context.Context, v any) (content.Content, error) {
	b := content.NewBuilder(t.media)

	b.SetReadCloser(content.AsReadCloser(ctx, func(w io.WriteCloser) func() error {
		mw := multipart.NewWriter(w)
		b.SetContentType(mw.FormDataContentType())

		return func() error {
			defer func() { _ = w.Close() }()

			rv, ok := v.(reflect.Value)
			if ok {
				v = rv.Interface()
			}

			for rv.Kind() == reflect.Pointer {
				rv = rv.Elem()
			}

			pv := NewArshaler(rv)
			s, err := scanner.Structs.Scan(pv.Type())
			if err != nil {
				return err
			}

			for sf := range s.Range {
				for sfv := range pv.Values(sf) {
					if sfv.IsZero() {
						if sf.Omitzero || sf.Omitempty {
							continue
						}
					}
					params := map[string]string{"name": sf.Name}
					fv := sfv.Interface()
					if x, ok := fv.(content.WithFilename); ok {
						params["filename"] = x.Filename()
					}

					header := textproto.MIMEHeader{}
					header.Set("Content-Disposition", mime.FormatMediaType("form-data", params))

					tf, err := New(sfv.Type(), sf.Tag.Get("mime"), ForMarshalling)
					if err != nil {
						return err
					}
					c, err := tf.Prepare(ctx, sfv)
					if err != nil {
						return err
					}
					if ct := c.ContentType(); ct != "" {
						header.Set("Content-Type", ct)
					}
					if x, ok := fv.(content.MediaTypeDescriber); ok {
						header.Set("Content-Type", x.ContentType())
					}
					if n := c.ContentLength(); n > -1 {
						header.Set("Content-Length", strconv.FormatInt(n, 10))
					}
					var pw io.Writer
					pw, err = mw.CreatePart(header)
					if err != nil {
						return err
					}
					if _, err = io.Copy(pw, c); err != nil {
						return err
					}

				}
			}
			return mw.Close()
		}
	}))

	return b, nil
}
