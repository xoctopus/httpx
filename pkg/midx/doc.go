// Package midx provides HTTP middleware factories for [httpx] applications.
//
// Each exported constructor returns a standard net/http middleware (or
// [httpx.Middleware] where noted). Compose them with [httpx.ApplyMiddlewares]
// so the first argument runs outermost on the request path.
//
// Example:
//
//	h, err := routex.NewHandler(apis.R, "example")
//	if err != nil {
//		panic(err)
//	}
//
//	h = httpx.ApplyMiddlewares(
//		midx.InjectContext(user.Carry(user.New())),
//		midx.LogHandler(),
//		midx.DefaultCORS(),
//	)(h)
//
//	if err = confhttp.ListenAndServe(ctx, "0.0.0.0:9001", h); err != nil {
//		panic(err)
//	}
//
// # Middleware
//
//   - [InjectContext]: compose [contextx.Carrier] values into request context
//     before the handler runs.
//   - [LogHandler]: access logging with B3 trace propagation, optional span from
//     [httpx.OperationMetaFrom], and per-request level override via the
//     x-enable-log-level header.
//   - [MetricHandler]: serve Prometheus metrics at /.sys/metrics.
//   - [DefaultCORS] / [CORS]: Cross-Origin Resource Sharing; [DefaultCORS] sets
//     httpx-friendly defaults (methods, headers including Authorization, exposed
//     rate-limit headers, credentials).
//   - [CanonicalHost]: redirect to a canonical host/scheme (from gorilla/handlers).
//   - [DefaultCompress] / [Compress]: gzip and brotli response compression when
//     Accept-Encoding allows it.
//   - [PProfHandler]: optional net/http/pprof under /.sys/debug/pprof when enabled.
//
// # Internal packages
//
// Implementation details live under internal/; midx re-exports thin wrappers and
// functional options where applicable.
//
//   - logginghandler: [ResponseWriter] wrapper that records status, treats 4xx/5xx
//     bodies as errors for logs, and [OmitAuthorization] for safe URL logging.
//   - corshandler: CORS header handling and preflight (from gorilla/handlers).
//   - canonicalhandler: host/scheme redirect handler (from gorilla/handlers).
//   - compresshandler: gzip/brotli encoding (from gorilla/handlers).
//   - pprofhandler: path-prefixed pprof endpoints under /.sys/debug/pprof.
package midx
