package content

import (
	"context"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/xoctopus/logx"
)

func New(typ string) Content {
	return &content{typ: typ, len: -1}
}

func NewFrom(media string, param map[string]string) Content {
	return New(mime.FormatMediaType(media, param))
}

func NewBuilder() Builder {
	return &content{len: -1}
}

func NewBuilderFrom(media string, param map[string]string) Builder {
	return &content{typ: mime.FormatMediaType(media, param)}
}

type content struct {
	typ string
	len int64
	io.ReadCloser
}

func (c content) Type() string {
	return c.typ
}

func (c content) ContentType() string {
	return c.typ
}

func (c *content) SetContentType(typ string) {
	c.typ = typ
}

func (c content) Length() int64 {
	return c.len
}

func (c *content) SetLength(n int64) {
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
	var (
		pr, pw = io.Pipe()
		write  = factory(pw)
	)

	go func() {
		if err := write(); err != nil {
			logx.From(ctx).Error(err)
		}
	}()

	return pr
}
