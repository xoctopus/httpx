package midx

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricHandler(gatherer prometheus.Gatherer) func(http.Handler) http.Handler {
	h := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/.sys/metrics") {
				h.ServeHTTP(rw, req)
				return
			}
			handler.ServeHTTP(rw, req)
		})
	}
}
