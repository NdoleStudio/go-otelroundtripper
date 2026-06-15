package otelroundtripper

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestNew(t *testing.T) {
	t.Run("with no options", func(t *testing.T) {
		t.Parallel()
		roundTripper := New()
		assert.NotNil(t, roundTripper)
	})

	t.Run("with options", func(t *testing.T) {
		t.Parallel()
		roundTripper := New(
			WithName("name"),
			WithParent(http.DefaultTransport),
			WithMeter(noop.NewMeterProvider().Meter("http.client")),
			WithAttributes([]attribute.KeyValue{semconv.ServiceNameKey.String("service")}...),
		)
		assert.NotNil(t, roundTripper)
	})
}

func TestOtelRoundTripper_RoundTrip(t *testing.T) {
	// Setup
	t.Parallel()
	server := makeTestServer(http.StatusOK, http.StatusText(http.StatusOK), 0)

	// Arrange
	client := &http.Client{
		Transport: New(),
	}

	// Act
	response, err := client.Get(server.URL)

	// Assert
	assert.Nil(t, err)

	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusText(http.StatusOK), string(body))

	// Teardown
	server.Close()
}

func TestOtelRoundTripper_RoundTripWithTimeout(t *testing.T) {
	// Setup
	t.Parallel()
	server := makeTestServer(http.StatusOK, http.StatusText(http.StatusOK), 100)

	// Arrange
	client := &http.Client{
		Transport: New(),
		Timeout:   10,
	}

	// Act
	_, err := client.Get(server.URL) //nolint:bodyclose

	// Assert
	assert.NotNil(t, err)

	var timeoutError net.Error
	assert.True(t, errors.As(err, &timeoutError) && timeoutError.Timeout())

	server.Close()
}

func TestOtelRoundTripper_RoundTripWithCancelledContext(t *testing.T) {
	// Setup
	t.Parallel()
	server := makeTestServer(http.StatusOK, http.StatusText(http.StatusOK), 200)
	defer server.Close()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { _ = provider.Shutdown(context.Background()) }()

	client := &http.Client{
		Transport: New(
			WithName("test"),
			WithMeter(provider.Meter("test")),
		),
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	assert.Nil(t, err)
	cancelFunc()

	_, err = client.Do(request) //nolint:bodyclose

	// Assert
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.Canceled))

	// Collect all metrics from the SDK
	var rm metricdata.ResourceMetrics
	assert.Nil(t, reader.Collect(context.Background(), &rm))

	var canceledSum int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "test.cancelled" {
				data, ok := m.Data.(metricdata.Sum[int64])
				if ok {
					for _, dp := range data.DataPoints {
						canceledSum += dp.Value
					}
				}
			}
		}
	}
	assert.Equal(t, int64(1), canceledSum, "expected cancelled counter to be 1")
}

func TestOtelRoundTripper_RoundTripWithDeadlineExceeded(t *testing.T) {
	// Setup
	t.Parallel()
	server := makeTestServer(http.StatusOK, http.StatusText(http.StatusOK), 200)
	defer server.Close()

	// Arrange: use a real SDK meter with a manual reader so we can inspect emitted metrics
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { _ = provider.Shutdown(context.Background()) }()

	client := &http.Client{
		Transport: New(
			WithName("test"),
			WithMeter(provider.Meter("test")),
		),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	assert.Nil(t, err)

	// Act: request will hit the slow server and the context deadline will expire
	_, err = client.Do(req) //nolint:bodyclose
	assert.NotNil(t, err)

	// Collect all metrics from the SDK
	var rm metricdata.ResourceMetrics
	assert.Nil(t, reader.Collect(context.Background(), &rm))

	// Find the deadline_exceeded counter and assert it was incremented
	var deadlineExceededSum int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "test.deadline_exceeded" {
				data, ok := m.Data.(metricdata.Sum[int64])
				if ok {
					for _, dp := range data.DataPoints {
						deadlineExceededSum += dp.Value
					}
				}
			}
		}
	}
	assert.Equal(t, int64(1), deadlineExceededSum, "expected deadline_exceeded counter to be 1")
}

func TestOtelRoundTripper_RoundTripWithNilRequest(t *testing.T) {
	t.Parallel()

	// use a real SDK meter with a manual reader so it can inspect emitted metrics
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { _ = provider.Shutdown(context.Background()) }()

	// Create a custom parent transport that handles nil requests
	customParent := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if req == nil {
				return nil, errors.New("nil request")
			}
			return &http.Response{StatusCode: http.StatusOK}, nil
		},
	}

	roundTripper := New(
		WithName("test"),
		WithMeter(provider.Meter("test")),
		WithParent(customParent),
	).(*otelRoundTripper)

	//call RoundTrip with nil request
	_, _ = roundTripper.RoundTrip(nil) //nolint:bodyclose

	// Collect all metrics from the SDK
	var rm metricdata.ResourceMetrics
	assert.Nil(t, reader.Collect(context.Background(), &rm))

	// Find the no_request counter and assert it was incremented
	var noRequestSum int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "test.no_request" {
				data, ok := m.Data.(metricdata.Sum[int64])
				if ok {
					for _, dp := range data.DataPoints {
						noRequestSum += dp.Value
					}
				}
			}
		}
	}
	assert.Equal(t, int64(1), noRequestSum, "expected no_request counter to be 1")
}

type mockRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

// makeTestServer creates an api server for testing
func makeTestServer(responseCode int, body string, delay int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(responseCode)

		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		_, err := res.Write([]byte(body))
		if err != nil {
			panic(err)
		}
	}))
}
