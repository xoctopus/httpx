package confhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/xoctopus/x/urlx"

	"github.com/xoctopus/httpx/internal/client"
	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/payload/metadata"
	"github.com/xoctopus/httpx/internal/payload/transformer"
	"github.com/xoctopus/httpx/internal/request"
	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/transport"
	"github.com/xoctopus/httpx/pkg/httpx"
)

type Transport = func(rt http.RoundTripper) http.RoundTripper

type Client struct {
	Endpoint  string `url:""`
	EnableH2C bool   `url:",omitzero"`

	NewError   func() error `url:"-"`
	transports []Transport  `url:"-"`
	endpoint   *url.URL
}

func (c *Client) Init(ctx context.Context) error {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return err
	}
	u2 := urlx.From(*u, urlx.WithScheme("http"))
	c.endpoint = &u2.URL
	return nil
}

func (c *Client) WithTransports(transports ...Transport) *Client {
	c2 := *c
	c2.transports = transports
	return &c2
}

func (c *Client) Do(ctx context.Context, vreq any, metas ...httpx.Metadata) httpx.Result {
	req, ok := vreq.(*http.Request)
	if !ok {
		r, err := c.newreq(ctx, vreq, metas...)
		if err != nil {
			return &result{
				client: c,
				err:    status.Wrap(err, httpx.STATUS__INTERNAL_SERVER_ERROR),
			}
		}
		req = r
	}

	hc := MustHttpClient(ctx)
	if hc == nil {
		hc = DefaultClientContext(ctx)
	}
	if hc.Transport == nil {
		hc.Transport = gDefaultRoundTripper
	}
	if c.EnableH2C {
		hc.Transport = client.H2CTransport(hc.Transport)
	}
	hc.Transport = WithHttpTransports(c.transports...)(hc.Transport)

	rsp, err := hc.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return &result{
				client: c,
				err:    status.Wrap(err, httpx.STATUS__CLIENT_CLOSED_REQUEST),
			}
		}
		return &result{
			client: c,
			err:    status.Wrap(err, httpx.STATUS__INTERNAL_SERVER_ERROR),
		}
	}

	return &result{
		client:   c,
		Response: rsp,
	}
}

func (c *Client) newreq(ctx context.Context, r any, metas ...httpx.Metadata) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := transport.NewRequest(ctx, r)
	if err != nil {
		return nil, status.Wrap(
			fmt.Errorf("failed to new request: %w", err),
			httpx.STATUS__BAD_REQUEST,
		)
	}

	// as default endpoint schema, host and path
	req.URL.Scheme = c.endpoint.Scheme
	req.URL.Host = c.endpoint.Host
	req.URL.Path = path.Clean(c.endpoint.Path + req.URL.Path)

	for k, vs := range httpx.MergeMetadata(metas...) {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	return req, nil
}

type result struct {
	*http.Response
	client *Client
	err    error
}

func (r *result) Into(v any) (httpx.Metadata, error) {
	autoClose := true

	defer func() {
		if autoClose {
			if r.Response != nil && r.Response.Body != nil {
				_ = r.Response.Body.Close()
			}
		}
	}()

	if r.err != nil {
		return nil, r.err
	}

	meta := metadata.Metadata(r.Response.Header)

	if !httpx.IsStatusOK(r.Response.StatusCode) {
		if r.client.NewError != nil {
			v = r.client.NewError()
		} else {
			v = &status.Description{Source: r.Response.Request.Host}
		}
	}

	if v == nil {
		return meta, nil
	}

	switch x := v.(type) {
	case *io.ReadCloser:
		autoClose = false
		*x = r.Response.Body
		return meta, nil
	case *any:
		return meta, nil
	case error:
		switch u := x.(type) {
		case status.CanUnmarshalResponse:
			if r.Response != nil && r.Response.Body != nil {
				data, err := io.ReadAll(r.Response.Body)
				if err != nil {
					return meta, err
				}
				if err = u.UnmarshalResponse(r.Response.StatusCode, data); err != nil {
					return nil, err
				}
				return meta, x
			}
			if err := u.UnmarshalResponse(r.Response.StatusCode, nil); err != nil {
				return nil, err
			}
			return meta, x
		default:
			if err := r.into(u); err != nil {
				return meta, err
			}
			return meta, x
		}
	case io.Writer:
		if _, err := io.Copy(x, r.Response.Body); err != nil {
			err = fmt.Errorf("WriteResponseFailed: %w", err)
			return meta, status.Wrap(err, httpx.STATUS__INTERNAL_SERVER_ERROR)
		}
	default:
		if err := r.into(v); err != nil {
			return meta, err
		}
	}

	return meta, nil
}

func (r *result) into(v any) error {
	rv := reflect.ValueOf(v)

	media := strings.Split(r.Response.Header.Get("Content-Type"), ";")[0]
	if x, ok := v.(content.MediaTypeDescriber); ok {
		media = x.ContentType()
	}

	f, err := transformer.New(rv.Type(), media, transformer.ForUnmarshalling)
	if err != nil {
		return err
	}
	if err = f.Into(context.Background(), request.ReadCloserWithHeader(r.Response.Body, r.Response.Header), rv); err != nil {
		err = fmt.Errorf("ResponseDecodeFailed: unmarshal to %T[%w]", v, err)
		return status.Wrap(err, httpx.STATUS__INTERNAL_SERVER_ERROR)
	}

	return nil
}
