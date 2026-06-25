package middlex

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xoctopus/logx"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"

	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/middlex/internal/logginghandler"
)

func LogHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			var (
				ctx     = req.Context()
				ts      = time.Now()
				lrw     = logginghandler.NewResponseWriter(rw)
				info, _ = httpx.OperationMetaFrom(ctx)
			)

			ctx = b3.New().Extract(ctx, propagation.HeaderCarrier(req.Header))
			b3.New().Inject(ctx, propagation.HeaderCarrier(lrw.Header()))

			if info != nil {
				var span logx.Logger
				ctx, span = logx.From(ctx).Start(ctx, info.ID)
				defer span.End()
			}

			next.ServeHTTP(lrw, req.WithContext(ctx))
			cost := time.Since(ts)

			lvl := logx.LogLevelInfo
			if v := req.Header.Get("x-enable-log-level"); v != "" {
				if x, ok := logginghandler.LogLevels[strings.ToLower(v)]; ok {
					lvl = x
				}
			}

			log := logx.From(ctx).With(
				"http.client_ip", httpx.ClientIP(req),
				"http.method", req.Method,
				"http.proto", req.Proto,
				"http.url", logginghandler.OmitAuthorization(req.URL),
				"http.status", lrw.StatusCode(),
				"http.ua", req.Header.Get("User-Agent"),
				"http.srv.cost(ms)", cost.Milliseconds(),
			)

			if err := lrw.StatusError(); err != nil {
				if lrw.StatusCode() >= http.StatusInternalServerError {
					log.Error(err)
				} else {
					if logx.LogLevelWarn >= lvl {
						log.Warn(err)
					}
				}
			} else {
				if logx.LogLevelInfo >= lvl {
					log.Info("success")
				}
			}
		})
	}
}

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
