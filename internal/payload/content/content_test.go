package content_test

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/go-json-experiment/json"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/payload/content"
)

func TestContent(t *testing.T) {
	c := content.NewFrom("application/json", map[string]string{"charset": "utf-8"})
	Expect(t, c.ContentType(), Equal("application/json; charset=utf-8"))

	b := content.NewBuilderFrom("application/json", map[string]string{"charset": "utf-8"})
	Expect(t, b.ContentType(), Equal("application/json; charset=utf-8"))
	b.SetContentType("application/x-www-form-urlencoded")
	Expect(t, b.ContentType(), Equal("application/x-www-form-urlencoded"))

	type T struct {
		A int `json:"a"`
	}

	var (
		in     = T{A: 100}
		out    T
		length int64
	)

	// preparing
	b.SetReadCloser(content.AsReadCloser(
		context.Background(),
		func(w io.WriteCloser) func() error {
			return func() error {
				defer func() { _ = w.Close() }()
				return json.MarshalWrite(w, in)
			}
		},
	))

	defer func() { _ = b.Close() }()
	err := content.Pipe(
		func(r io.Reader) error {
			return json.UnmarshalRead(r, &out)
		},
		func(w io.Writer) error {
			data, _ := io.ReadAll(b)
			_, _ = w.Write(data)
			length = int64(len(data))
			b.SetContentLength(length)
			return nil
		},
	)

	Expect(t, err, Succeed())
	Expect(t, length, Equal(b.ContentLength()))

	applier, header := b.(content.HeaderApplier), http.Header{}
	applier.ApplyHeader(header)
	Expect(t, header["Content-Type"][0], Equal(b.ContentType()))
	Expect(t, header["Content-Length"][0], Equal(strconv.FormatInt(b.ContentLength(), 10)))
}
