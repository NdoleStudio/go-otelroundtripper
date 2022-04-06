package otelroundtripper

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type otelHTTPMetrics struct {
	attemptsCounter         syncint64.Counter
	noRequestCounter        syncint64.Counter
	errorsCounter           syncint64.Counter
	successesCounter        syncint64.Counter
	failureCounter          syncint64.Counter
	redirectCounter         syncint64.Counter
	timeoutsCounter         syncint64.Counter
	canceledCounter         syncint64.Counter
	deadlineExceededCounter syncint64.Counter
	totalDurationCounter    syncint64.Histogram
	inFlightCounter         syncint64.UpDownCounter
}

// otelRoundTripper is the http.RoundTripper which emits open telemetry metrics
type otelRoundTripper struct {
	parent     http.RoundTripper
	attributes []attribute.KeyValue
	metrics    otelHTTPMetrics
}

// New creates a new instance of the http.RoundTripper
func New(options ...Option) http.RoundTripper {
	cfg := defaultConfig()

	for _, option := range options {
		option.apply(cfg)
	}

	return &otelRoundTripper{
		parent:     cfg.parent,
		attributes: cfg.attributes,
		metrics: otelHTTPMetrics{
			noRequestCounter:        mustCounter(cfg.meter.SyncInt64().Counter(cfg.name + ".no_request")),
			errorsCounter:           mustCounter(cfg.meter.SyncInt64().Counter(cfg.name + ".errors")),
			successesCounter:        mustCounter(cfg.meter.SyncInt64().Counter(cfg.name + ".success")),
			timeoutsCounter:         mustCounter(cfg.meter.SyncInt64().Counter(cfg.name + ".timeouts")),
			canceledCounter:         mustCounter(cfg.meter.SyncInt64().Counter(cfg.name + ".cancelled")),
			deadlineExceededCounter: mustCounter(cfg.meter.SyncInt64().Counter(cfg.name + ".deadline_exceeded")),
			totalDurationCounter:    mustHistogram(cfg.meter.SyncInt64().Histogram(cfg.name + ".total_duration")),
			inFlightCounter:         mustUpDownCounter(cfg.meter.SyncInt64().UpDownCounter(cfg.name + ".in_flight")),
			attemptsCounter:         mustCounter(cfg.meter.SyncInt64().Counter(cfg.name + ".attempts")),
			failureCounter:          mustCounter(cfg.meter.SyncInt64().Counter(cfg.name + ".failures")),
			redirectCounter:         mustCounter(cfg.meter.SyncInt64().Counter(cfg.name + ".redirects")),
		},
	}
}

func mustCounter(counter syncint64.Counter, err error) syncint64.Counter {
	if err != nil {
		panic(err)
	}
	return counter
}

func mustUpDownCounter(counter syncint64.UpDownCounter, err error) syncint64.UpDownCounter {
	if err != nil {
		panic(err)
	}
	return counter
}

func mustHistogram(histogram syncint64.Histogram, err error) syncint64.Histogram {
	if err != nil {
		panic(err)
	}
	return histogram
}

// RoundTrip executes a single HTTP transaction, returning a Response for the provided Request.
func (roundTripper *otelRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx := roundTripper.extractCtx(request)
	attributes := roundTripper.requestAttributes(request)

	roundTripper.beforeHook(ctx, attributes, request)

	start := time.Now()
	response, err := roundTripper.parent.RoundTrip(request)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		roundTripper.errorHook(ctx, err, attributes)
		return response, err
	}

	attributes = roundTripper.responseAttributes(attributes, response)
	roundTripper.afterHook(ctx, duration, attributes)

	if roundTripper.isRedirection(response) {
		roundTripper.redirectHook(ctx, attributes)
		return response, err
	}

	if roundTripper.isFailure(response) {
		roundTripper.failureHook(ctx, attributes)
		return response, err
	}

	roundTripper.successHook(ctx, attributes)
	return response, err
}

func (roundTripper *otelRoundTripper) isFailure(response *http.Response) bool {
	if response == nil {
		return false
	}
	return response.StatusCode >= http.StatusBadRequest
}

func (roundTripper *otelRoundTripper) isRedirection(response *http.Response) bool {
	if response == nil {
		return false
	}
	return response.StatusCode >= http.StatusMultipleChoices && response.StatusCode < http.StatusBadRequest
}

func (roundTripper *otelRoundTripper) failureHook(
	ctx context.Context,
	attributes []attribute.KeyValue,
) {
	roundTripper.metrics.inFlightCounter.Add(ctx, -1, attributes...)
	roundTripper.metrics.failureCounter.Add(ctx, 1, attributes...)
}

func (roundTripper *otelRoundTripper) redirectHook(
	ctx context.Context,
	attributes []attribute.KeyValue,
) {
	roundTripper.metrics.inFlightCounter.Add(ctx, -1, attributes...)
	roundTripper.metrics.redirectCounter.Add(ctx, 1, attributes...)
}

func (roundTripper *otelRoundTripper) successHook(
	ctx context.Context,
	attributes []attribute.KeyValue,
) {
	roundTripper.metrics.inFlightCounter.Add(ctx, -1, attributes...)
	roundTripper.metrics.successesCounter.Add(ctx, 1, attributes...)
}

func (roundTripper *otelRoundTripper) beforeHook(
	ctx context.Context,
	attributes []attribute.KeyValue,
	request *http.Request,
) {
	roundTripper.metrics.inFlightCounter.Add(ctx, 1, attributes...)
	roundTripper.metrics.attemptsCounter.Add(ctx, 1, attributes...)
	if request == nil {
		roundTripper.metrics.noRequestCounter.Add(ctx, 1, attributes...)
	}
}

func (roundTripper *otelRoundTripper) afterHook(
	ctx context.Context,
	duration int64,
	attributes []attribute.KeyValue,
) {
	roundTripper.metrics.totalDurationCounter.Record(ctx, duration, attributes...)
}

func (roundTripper *otelRoundTripper) responseAttributes(
	attributes []attribute.KeyValue,
	response *http.Response,
) []attribute.KeyValue {
	return append(
		append([]attribute.KeyValue{}, attributes...),
		roundTripper.extractResponseAttributes(response)...,
	)
}

func (roundTripper *otelRoundTripper) requestAttributes(request *http.Request) []attribute.KeyValue {
	return append(
		append(
			[]attribute.KeyValue{},
			roundTripper.attributes...,
		),
		roundTripper.extractRequestAttributes(request)...,
	)
}

func (roundTripper *otelRoundTripper) errorHook(ctx context.Context, err error, attributes []attribute.KeyValue) {
	roundTripper.metrics.inFlightCounter.Add(ctx, -1, attributes...)
	roundTripper.metrics.errorsCounter.Add(ctx, 1, attributes...)

	var timeoutErr net.Error
	if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
		roundTripper.metrics.timeoutsCounter.Add(ctx, 1, attributes...)
	}

	if strings.HasSuffix(err.Error(), context.Canceled.Error()) {
		roundTripper.metrics.canceledCounter.Add(ctx, 1, attributes...)
	}
}

func (roundTripper *otelRoundTripper) extractResponseAttributes(response *http.Response) []attribute.KeyValue {
	if response != nil {
		return []attribute.KeyValue{
			semconv.HTTPStatusCodeKey.Int(response.StatusCode),
			semconv.HTTPFlavorKey.String(response.Proto),
		}
	}
	return nil
}

func (roundTripper *otelRoundTripper) extractRequestAttributes(request *http.Request) []attribute.KeyValue {
	if request != nil {
		return []attribute.KeyValue{
			semconv.HTTPURLKey.String(request.URL.String()),
			semconv.HTTPMethodKey.String(request.Method),
		}
	}
	return nil
}

func (roundTripper *otelRoundTripper) extractCtx(request *http.Request) context.Context {
	if request != nil && request.Context() != nil {
		return request.Context()
	}
	return context.Background()
}
