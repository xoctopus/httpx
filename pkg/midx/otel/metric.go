package otel

import (
	"github.com/xoctopus/confx/pkg/confotel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/semconv/v1.43.0/httpconv"
)

// DurationHistogramBoundaries histogram boundaries for http request cost (seconds)
var DurationHistogramBoundaries = []float64{
	0,
	0.005,
	0.01,
	0.025,
	0.05,
	0.075,
	0.1,
	0.25,
	0.5,
	0.75,
	1,
	2.5,
	5,
	7.5,
	10,
}

const (
	B   = 1
	KiB = 1024 * B
	MiB = 1024 * KiB
)

// SizeHistogramBoundaries histogram boundaries for http request/response body size
var SizeHistogramBoundaries = []float64{
	512 * B,
	1 * KiB,
	64 * KiB,
	128 * KiB,
	256 * KiB,
	512 * KiB,
	1 * MiB,
	2 * MiB,
	5 * MiB,
	10 * MiB,
	20 * MiB,
	50 * MiB,
}

var (
	clientDuration    = httpconv.ClientRequestDuration{}
	clientReqBodySize = httpconv.ClientRequestBodySize{}
	clientRspBodySize = httpconv.ClientResponseBodySize{}

	serverDuration       = httpconv.ServerRequestDuration{}
	serverActiveRequests = httpconv.ServerActiveRequests{}
	serverReqBodySize    = httpconv.ServerRequestBodySize{}
	serverRspBodySize    = httpconv.ServerResponseBodySize{}
)

var (
	ClientDuration = metric.NewFloat64Histogram(
		clientDuration.Name(),
		metric.WithUnit(clientDuration.Unit()),
		metric.WithDescription(clientDuration.Description()),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: DurationHistogramBoundaries,
		}),
	)
	ClientRequestSize = metric.NewInt64Histogram(
		clientReqBodySize.Name(),
		metric.WithUnit(clientReqBodySize.Unit()),
		metric.WithDescription(clientReqBodySize.Description()),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)
	ClientResponseSize = metric.NewInt64Histogram(
		clientRspBodySize.Name(),
		metric.WithUnit(clientRspBodySize.Unit()),
		metric.WithDescription(clientRspBodySize.Description()),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)
)

var (
	ServerDuration = metric.NewFloat64Histogram(
		serverDuration.Name(),
		metric.WithUnit(serverDuration.Unit()),
		metric.WithDescription(serverDuration.Description()),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: DurationHistogramBoundaries,
		}),
	)
	ServerActiveRequests = metric.NewInt64UpDownCounter(
		serverActiveRequests.Name(),
		metric.WithDescription(serverActiveRequests.Description()),
	)
	ServerRequestSize = metric.NewInt64Histogram(
		serverReqBodySize.Name(),
		metric.WithUnit(serverReqBodySize.Unit()),
		metric.WithDescription(serverReqBodySize.Description()),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)
	ServerResponseSize = metric.NewInt64Histogram(
		serverRspBodySize.Name(),
		metric.WithUnit(serverRspBodySize.Unit()),
		metric.WithDescription(serverRspBodySize.Description()),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)
)
