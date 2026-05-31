package content

import (
	"context"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/xoctopus/logx"
	"golang.org/x/sync/errgroup"
)

func New(typ string) Content {
	return &content{typ: typ, len: -1}
}

func NewFrom(media string, param map[string]string) Content {
	return New(mime.FormatMediaType(media, param))
}

func NewBuilder(media string) Builder {
	return &content{typ: media, len: -1}
}

func NewBuilderFrom(media string, param map[string]string) Builder {
	return NewBuilder(mime.FormatMediaType(media, param))
}

type content struct {
	typ string
	len int64
	io.ReadCloser
}

func (c content) ContentType() string {
	return c.typ
}

func (c *content) SetContentType(t string) {
	c.typ = t
}

func (c content) ContentLength() int64 {
	return c.len
}

func (c *content) SetContentLength(n int64) {
	c.len = n
}

func (c *content) SetReadCloser(rc io.ReadCloser) {
	c.ReadCloser = rc
}

func (c *content) ApplyHeader(h http.Header) {
	if len(c.typ) > 0 {
		h.Set("Content-Type", c.typ)
	}
	if c.len > -1 {
		h.Set("Content-Length", strconv.FormatInt(c.len, 10))
	}
}

func AsReadCloser(ctx context.Context, factory func(w io.WriteCloser) func() error) io.ReadCloser {
	pr, pw := io.Pipe()
	w := factory(pw)

	go func() {
		if err := w(); err != nil {
			logx.From(ctx).Error(err)
		}
	}()

	return pr
}

func Pipe(r func(r io.Reader) error, w func(w io.Writer) error) error {
	pr, pw := io.Pipe()

	eg := &errgroup.Group{}

	eg.Go(func() (err error) {
		defer func() {
			_ = pr.CloseWithError(err)
		}()
		return r(pr)
	})

	eg.Go(func() (err error) {
		defer func() {
			_ = pw.CloseWithError(err)
		}()
		return w(pw)
	})

	return eg.Wait()
}
