module github.com/xoctopus/httpx

go 1.27.0

tool (
	github.com/xoctopus/httpx/internal/cmd/gen
	github.com/xoctopus/httpx/internal/cmd/skill-install
)

require (
	github.com/andybalholm/brotli v1.2.6
	github.com/fatih/color v1.19.0
	github.com/felixge/httpsnoop v1.1.0
	github.com/prometheus/client_golang v1.25.0
	// +skill:genx
	github.com/xoctopus/genx v0.4.0
	// +skill:logx
	github.com/xoctopus/logx v0.3.9
	// +skill:testx
	github.com/xoctopus/x v0.5.9
	go.opentelemetry.io/contrib/propagators/b3 v1.47.0
	go.opentelemetry.io/otel v1.47.0
	golang.org/x/sync v0.24.0
	google.golang.org/protobuf v1.36.12
	k8s.io/apimachinery v0.37.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/xoctopus/pkgx v0.4.4 // indirect
	github.com/xoctopus/typx v0.4.7 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
)
