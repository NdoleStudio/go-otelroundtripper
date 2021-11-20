package otelroundtripper

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Option applies a configuration to the given config
type Option interface {
	apply(cfg *config)
}

type optionFunc func(cfg *config)

func (fn optionFunc) apply(cfg *config) {
	fn(cfg)
}

// WithParent sets the underlying http.RoundTripper which is wrapped by this round tripper.
// If the provided http.RoundTripper is nil, http.DefaultTransport will be used as the base http.RoundTripper
func WithParent(parent http.RoundTripper) Option {
	return optionFunc(func(cfg *config) {
		if parent != nil {
			cfg.parent = parent
		}
	})
}

// WithName sets the prefix for the metrics emitted by this round tripper.
// by default, the "http.client" name is used.
func WithName(name string) Option {
	return optionFunc(func(cfg *config) {
		if strings.TrimSpace(name) != "" {
			cfg.name = strings.TrimSpace(name)
		}
	})
}

// WithMeter sets the underlying  metric.Meter that is used to create metric instruments
// By default the no-op meter is used.
func WithMeter(meter metric.Meter) Option {
	return optionFunc(func(cfg *config) {
		cfg.meter = meter
	})
}

// WithAttributes sets a list of attribute.KeyValue labels for all metrics associated with this round tripper
func WithAttributes(attributes ...attribute.KeyValue) Option {
	return optionFunc(func(cfg *config) {
		cfg.attributes = attributes
	})
}
