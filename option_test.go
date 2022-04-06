package otelroundtripper

import (
	"net/http"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/nonrecording"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/stretchr/testify/assert"
)

func TestWithParent(t *testing.T) {
	t.Run("parent is not set when the parent is nil", func(t *testing.T) {
		// Setup
		t.Parallel()

		// Arrange
		cfg := defaultConfig()

		// Act
		WithParent(nil).apply(cfg)

		// Assert
		assert.NotNil(t, cfg.parent)
	})

	t.Run("parent is set when the parent is not nil", func(t *testing.T) {
		// Setup
		t.Parallel()

		// Arrange
		cfg := defaultConfig()
		parent := &http.Transport{}

		// Act
		WithParent(parent).apply(cfg)

		// Assert
		assert.NotNil(t, cfg.parent)
		assert.Equal(t, parent, cfg.parent)
	})
}

func TestWithName(t *testing.T) {
	t.Run("name is not set when the parent is an empty string", func(t *testing.T) {
		// Setup
		t.Parallel()

		// Arrange
		cfg := defaultConfig()

		// Act
		WithName("   ").apply(cfg)

		// Assert
		assert.Equal(t, "http.client", cfg.name)
	})

	t.Run("name is set when the name is not an empty string", func(t *testing.T) {
		// Setup
		t.Parallel()

		// Arrange
		cfg := defaultConfig()
		name := " name "

		// Act
		WithName(name).apply(cfg)

		// Assert
		assert.Equal(t, "name", cfg.name)
	})
}

func TestWithMeter(t *testing.T) {
	t.Run("meter is set successfully", func(t *testing.T) {
		// Setup
		t.Parallel()

		// Arrange
		cfg := defaultConfig()
		meter := nonrecording.NewNoopMeter()

		// Act
		WithMeter(meter).apply(cfg)

		// Assert
		assert.Equal(t, meter, cfg.meter)
	})
}

func TestWithAttributes(t *testing.T) {
	t.Run("attributes are successfully", func(t *testing.T) {
		// Setup
		t.Parallel()

		// Arrange
		cfg := defaultConfig()
		attributes := []attribute.KeyValue{
			semconv.ServiceNamespaceKey.String("namespace"),
			semconv.ServiceInstanceIDKey.Int(1),
		}

		// Act
		WithAttributes(attributes...).apply(cfg)

		// Assert
		assert.Equal(t, attributes, cfg.attributes)
	})
}
