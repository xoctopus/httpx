package httpx

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/xoctopus/httpx/internal/transport"
)

type Transport interface {
	Serve(Router) error
}

func Run(router Router, transports ...Transport) {
	wg := &sync.WaitGroup{}

	for i := range transports {
		t := transports[i]
		wg.Go(func() {
			if err := t.Serve(router); err != nil {
				fmt.Println(err)
			}
		})
	}
	wg.Wait()
}

type (
	RoundTrip     func(r *http.Request) (*http.Response, error)
	HttpTransport = func(rt http.RoundTripper) http.RoundTripper

	Upgrader          = transport.Upgrader
	IncomingTransport = transport.Incoming
	OutgoingTransport = transport.Outgoing
)
