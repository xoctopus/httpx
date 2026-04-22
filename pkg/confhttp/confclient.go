package confhttp

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"path"

	"github.com/xoctopus/x/urlx"

	"github.com/xoctopus/httpx/internal/client"
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
				err:    status.WrapStatus(err, httpx.STATUS__INTERNAL_SERVER_ERROR),
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
				err:    status.WrapStatus(err, httpx.STATUS__CLIENT_CLOSED_REQUEST),
			}
		}
		return &result{
			client: c,
			err:    status.WrapStatus(err, httpx.STATUS__INTERNAL_SERVER_ERROR),
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
		return nil, status.Wrap(err, http.StatusBadRequest, "NewRequestFailed")
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
	// TODO
	return nil, nil
}
