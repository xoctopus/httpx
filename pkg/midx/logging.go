package midx

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xoctopus/x/logx"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/semconv/v1.43.0/httpconv"

	"github.com/xoctopus/httpx/internal/types"
	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/httpx/pkg/midx/internal/logginghandler"
	"github.com/xoctopus/httpx/pkg/midx/otel"
)

// Access log keys written by [LogHandler], follow OTel semconv attribute names.
// ref: https://opentelemetry.io/docs/specs/semconv/http/http-spans/#http-server
const (
	LogKeyClientAddress   = string(semconv.ClientAddressKey)
	LogKeyRequestMethod   = string(semconv.HTTPRequestMethodKey)
	LogKeyProtocolVersion = string(semconv.NetworkProtocolVersionKey)
	LogKeyURLPath         = string(semconv.URLPathKey)
	LogKeyURLQuery        = string(semconv.URLQueryKey)
	LogKeyUserAgent       = string(semconv.UserAgentOriginalKey)
	LogKeyStatusCode      = string(semconv.HTTPResponseStatusCodeKey)
)

// logKeyDuration reuses the metric name http.server.request.duration, value in seconds.
var logKeyDuration = (httpconv.ServerRequestDuration{}).Name()

func DefaultLogHandler() func(http.Handler) http.Handler {
	return LogHandler(false)
}

func LogHandler(enableOTel bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			var (
				ctx     = req.Context()
				info, _ = httpx.OperationMetaFrom(ctx)
				lrw     = logginghandler.NewResponseWriter(rw)
				attrs   []attribute.KeyValue
			)

			if info == nil {
				info = &types.OperationMeta{ID: "unknown"}
			}

			if enableOTel {
				ctx = b3.New().Extract(ctx, propagation.HeaderCarrier(req.Header))
				b3.New().Inject(ctx, propagation.HeaderCarrier(lrw.Header()))

				var span logx.Logger
				ctx, span = logx.From(ctx).Start(ctx, info.ID)
				defer span.End()

				attrs = ExtractBasicAttrs(req)
				otel.ServerActiveRequests.Add(ctx, 1, metric.WithAttributes(attrs...))
				defer otel.ServerActiveRequests.Add(ctx, -1, metric.WithAttributes(attrs...))
			}

			now := time.Now()
			next.ServeHTTP(lrw, req.WithContext(ctx))
			cost := time.Since(now)

			lvl := logx.LogLevelInfo
			if v := req.Header.Get("X-Enable-Log-Level"); v != "" {
				if x, ok := logginghandler.LogLevels[strings.ToLower(v)]; ok {
					lvl = x
				}
			}
			proto := strconv.Itoa(req.ProtoMajor)
			if req.ProtoMajor < 2 {
				proto += "." + strconv.Itoa(req.ProtoMinor)
			}
			log := logx.From(ctx).With(
				slog.String(LogKeyClientAddress, httpx.ClientIP(req)),
				slog.String(LogKeyRequestMethod, req.Method),
				slog.String(LogKeyProtocolVersion, proto),
				slog.String(LogKeyURLPath, req.URL.Path),
				slog.String(LogKeyURLQuery, logginghandler.OmitAuthorization(req.URL)),
				slog.String(LogKeyUserAgent, req.Header.Get("User-Agent")),
				slog.Int(LogKeyStatusCode, lrw.StatusCode()),
				slog.Float64(logKeyDuration, cost.Seconds()),
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

			if enableOTel {
				attrs = append(attrs, ExtractRouteAttrs(lrw.StatusCode(), info, req)...)
				otel.ServerDuration.Record(ctx, cost.Seconds(), metric.WithAttributes(attrs...))
				otel.ServerRequestSize.Record(ctx, req.ContentLength, metric.WithAttributes(attrs...))
				otel.ServerResponseSize.Record(ctx, lrw.ContentLength(), metric.WithAttributes(attrs...))
			}
		})
	}
}

func ExtractBasicAttrs(req *http.Request) []attribute.KeyValue {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}

	return []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String(req.Method),
		semconv.URLSchemeKey.String(scheme),
	}
}

func ExtractRouteAttrs(status int, info *types.OperationMeta, req *http.Request) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.HTTPRouteKey.String(info.Route),
		semconv.HTTPRequestMethodKey.String(req.Method),
		semconv.HTTPResponseStatusCodeKey.Int(status),
		semconv.NetworkProtocolNameKey.String(strings.ToLower(strings.Split(req.Proto, "/")[0])),
		semconv.NetworkProtocolVersionKey.String(fmt.Sprintf("%d.%d", req.ProtoMajor, req.ProtoMinor)),
	}
}
