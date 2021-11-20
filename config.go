package otelroundtripper

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type config struct {
	name       string
	parent     http.RoundTripper
	meter      metric.Meter
	attributes []attribute.KeyValue
}

var defaultConfig = &config{
	name:   "http.client",
	parent: http.DefaultTransport,
}
