package confhttp

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/httpx/internal/client"
)

type (
	tCtxClient              struct{}
	tCtxHttpClient          struct{}
	tCtxRoundTripperCreator struct{}
)

var (
	ClientFrom  = contextx.From[tCtxClient, client.Client]
	MustClient  = contextx.Must[tCtxClient, client.Client]
	WithClient  = contextx.With[tCtxClient, client.Client]
	CarryClient = contextx.Carry[tCtxClient, client.Client]

	HttpClientFrom  = contextx.From[tCtxHttpClient, *http.Client]
	MustHttpClient  = contextx.Must[tCtxHttpClient, *http.Client]
	WithHttpClient  = contextx.With[tCtxHttpClient, *http.Client]
	CarryHttpClient = contextx.Carry[tCtxHttpClient, *http.Client]

	RoundTripperCreatorFrom  = contextx.From[tCtxRoundTripperCreator, func() http.RoundTripper]
	WithRoundTripperCreator  = contextx.With[tCtxRoundTripperCreator, func() http.RoundTripper]
	MustRoundTripperCreator  = contextx.Must[tCtxRoundTripperCreator, func() http.RoundTripper]
	CarryRoundTripperCreator = contextx.Carry[tCtxRoundTripperCreator, func() http.RoundTripper]
)

type (
	DailContext   func(ctx context.Context, network string, address string) (net.Conn, error)
	HttpTransport = func(http.RoundTripper) http.RoundTripper
	RoundTrip     = func(*http.Request) (*http.Response, error)
)

var (
	gDefaultRoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: gDefaultHosts.WrapDialContext((&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext),
		MaxIdleConns:          100 * runtime.NumCPU(),
		MaxIdleConnsPerHost:   10 * runtime.NumCPU(),
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	gDefaultShortConnectionRoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: gDefaultHosts.WrapDialContext((&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 0,
		}).DialContext),
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}

	gDefaultHosts     = Hosts{}
	gDefaultTlsConfig = &tls.Config{}
)

func SetDefaultTLSClientConfig(c *tls.Config) {
	if c != nil {
		gDefaultTlsConfig = c.Clone()
		gDefaultRoundTripper.TLSClientConfig = c.Clone()
		gDefaultShortConnectionRoundTripper.TLSClientConfig = c.Clone()
	}
}

func AddHostAlias(aliases ...HostAlias) {
	for i := range len(aliases) {
		gDefaultHosts.AddHostAlias(aliases[i])
	}
}

func WithHttpTransports(transports ...HttpTransport) func(rt http.RoundTripper) http.RoundTripper {
	return func(r http.RoundTripper) http.RoundTripper {
		for _, t := range transports {
			r = t(r)
		}
		return r
	}
}

func DefaultClientContext(ctx context.Context, transports ...HttpTransport) *http.Client {
	t := http.RoundTripper(gDefaultRoundTripper)

	tc, ok := RoundTripperCreatorFrom(ctx)
	if ok {
		t = tc()
	}

	return &http.Client{Transport: WithHttpTransports(transports...)(t)}
}

func DefaultShortConnectionClientContext(ctx context.Context, transports ...HttpTransport) *http.Client {
	t := http.RoundTripper(gDefaultShortConnectionRoundTripper)

	tc, ok := RoundTripperCreatorFrom(ctx)
	if ok {
		t = tc()
	}

	return &http.Client{Transport: WithHttpTransports(transports...)(t)}
}

func HttpTransportFunc(round func(r *http.Request, next RoundTrip) (*http.Response, error)) HttpTransport {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &httpTransportFunc{
			rt:    rt,
			round: round,
		}
	}
}

type httpTransportFunc struct {
	rt    http.RoundTripper
	round func(*http.Request, RoundTrip) (*http.Response, error)
}

func (h *httpTransportFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return h.round(request, h.rt.RoundTrip)
}
