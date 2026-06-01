package response

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"

	"github.com/xoctopus/logx"

	"github.com/xoctopus/httpx/internal/client"
	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/payload/cookie"
	"github.com/xoctopus/httpx/internal/payload/metadata"
	"github.com/xoctopus/httpx/internal/payload/transformer"
	"github.com/xoctopus/httpx/internal/redirect"
	"github.com/xoctopus/httpx/internal/request"
	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/types"
)

type Writer interface {
	WriteResponse(context.Context, http.ResponseWriter, request.Request) error
}

type Response[T any] interface {
	Underlying() T

	content.MediaTypeDescriber
	status.Describer
	cookie.Describer
	metadata.Carrier
}

type Error interface {
	Error() string
	Unwrap() error

	content.MediaTypeDescriber
	status.Describer
	cookie.Describer
	metadata.Carrier
}

func New[T any](v T) Response[T] {
	return &response[T]{v: v}
}

type response[T any] struct {
	v        any
	metadata metadata.Metadata
	cookies  []*http.Cookie
	location *url.URL
	media    string
	status   int
}

func (r *response[T]) Underlying() T {
	return r.v.(T)
}

func (r *response[T]) Cookies() []*http.Cookie {
	return r.cookies
}

func (r *response[T]) SetStatusCode(code int) {
	r.status = code
}

func (r *response[T]) SetContentType(media string) {
	r.media = media
}

func (r *response[T]) SetMetadata(key string, values ...string) {
	if r.metadata == nil {
		r.metadata = map[string][]string{}
	}
	r.metadata[key] = values

	if r.media == "" || len(values) > 0 {
		if textproto.CanonicalMIMEHeaderKey(key) == "Content-Type" {
			r.media = values[0]
		}
	}
}

func (r *response[T]) SetCookies(cookies []*http.Cookie) {
	r.cookies = cookies
}

func (r *response[T]) SetLocation(location *url.URL) {
	r.location = location
}

func (r *response[T]) StatusCode() int {
	return r.status
}

func (r *response[T]) ContentType() string {
	return r.media
}

func (r *response[T]) Meta() metadata.Metadata {
	return r.metadata
}

func (r *response[T]) WriteResponse(ctx context.Context, rw http.ResponseWriter, req request.Request) (finalErr error) {
	defer func() {
		if x, ok := r.v.(io.Closer); ok {
			// close again to avoid some leak issue
			_ = x.Close()
		}
		r.v = nil
		if finalErr != nil {
			logx.From(ctx).Error(finalErr)
		}
	}()

	if w, ok := r.v.(Writer); ok {
		return w.WriteResponse(ctx, rw, req)
	}

	rsp := r.v

	if err, ok := r.v.(error); ok {
		meta, _ := types.OperationMetaFrom(ctx)
		rsp = status.AsResponse(err, meta.ServerMeta.UA())
	}

	if x, ok := rsp.(status.Describer); ok {
		r.SetStatusCode(x.StatusCode())
	}

	if r.status == 0 {
		if rsp == nil {
			r.SetStatusCode(http.StatusNoContent)
		} else {
			if req.Method() == http.MethodPost {
				r.SetStatusCode(http.StatusCreated)
			} else {
				r.SetStatusCode(http.StatusOK)
			}
		}
	}

	if r.location == nil {
		if x, ok := rsp.(redirect.Describer); ok {
			r.SetStatusCode(x.StatusCode())
			r.SetLocation(x.Location())
		}
	}

	if r.metadata != nil {
		header := rw.Header()
		for key, values := range r.metadata {
			if len(values) == 1 {
				if v := values[0]; len(v) > 0 && v[0] == ',' {
					if hv := header.Get(key); hv != "" {
						header.Set(key, hv+v)
						continue
					}
				}
			}
			header[textproto.CanonicalMIMEHeaderKey(key)] = values
		}
	}

	if r.cookies != nil {
		for i := range r.cookies {
			if ci := r.cookies[i]; ci != nil {
				http.SetCookie(rw, ci)
			}
		}
	}

	if r.location != nil {
		http.Redirect(rw, req.Underlying(), r.location.String(), r.status)
		return nil
	}

	if r.status == http.StatusNoContent {
		rw.WriteHeader(r.status)
		return nil
	}

	switch v := rsp.(type) {
	case client.Result:
		if r.media != "" {
			rw.Header().Set("Content-Type", r.media)
		}
		// forward result
		rw.WriteHeader(r.status)
		if _, err := v.Into(rw); err != nil {
			return fmt.Errorf("forward failed: %w", err)
		}
	default:
		if rsp == nil {
			// skip nil rsp
			rw.WriteHeader(r.status)
			return nil
		}

		t, err := transformer.New(reflect.TypeOf(rsp), "", transformer.ForMarshalling)
		if err != nil {
			return err
		}

		c, err := t.Prepare(ctx, rsp)
		if err != nil {
			return err
		}
		defer func() { _ = c.Close() }()

		if x, ok := c.(content.HeaderApplier); ok {
			x.ApplyHeader(rw.Header())
		}

		if r.media != "" {
			rw.Header().Set("Content-Type", r.media)
		}

		rw.WriteHeader(r.status)

		if _, err = io.Copy(rw, c); err != nil {
			return err
		}
	}

	return nil
}

type ErrorResponse struct {
	response[error]
}

func (e ErrorResponse) Error() string {
	return e.Underlying().Error()
}

func (e ErrorResponse) Unwrap() error {
	return e.Underlying()
}

type Modifier interface {
	status.Modifier
	redirect.Modifier
	content.MediaTypeModifier
	content.LengthModifier
	cookie.Modifier
	metadata.Modifier
}

type Applier func(m Modifier)
