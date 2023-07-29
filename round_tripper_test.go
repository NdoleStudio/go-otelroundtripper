package otelroundtripper

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/metric/noop"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
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

	// Teardown
	server.Close()
}

func TestOtelRoundTripper_RoundTripWithCancelledContext(t *testing.T) {
	// Setup
	t.Parallel()
	server := makeTestServer(http.StatusOK, http.StatusText(http.StatusOK), 0)

	// Arrange
	client := &http.Client{
		Transport: New(),
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	// Act
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	assert.Nil(t, err)

	_, err = client.Do(request) //nolint:bodyclose

	// Assert
	assert.NotNil(t, err)
	assert.True(t, strings.HasSuffix(err.Error(), context.Canceled.Error()))

	// Teardown
	server.Close()
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
