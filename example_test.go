package otelroundtripper

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func InstallExportPipeline(ctx context.Context) func() {
	exporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		log.Fatalf("creating stdoutmetric exporter: %v", err)
	}

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithInexpensiveDistribution(),
			exporter,
		),
		controller.WithExporter(exporter),
	)

	if err = pusher.Start(ctx); err != nil {
		log.Fatalf("starting push controller: %v", err)
	}

	global.SetMeterProvider(pusher)

	return func() {
		if err := pusher.Stop(ctx); err != nil {
			log.Fatalf("stopping push controller: %v", err)
		}
	}
}

func Example() {
	ctx := context.Background()

	// Registers a meter Provider globally.
	cleanup := InstallExportPipeline(ctx)
	defer cleanup()

	client := http.Client{
		Transport: New(
			WithMeter(global.Meter("otel-round-tripper")),
			WithAttributes(
				semconv.ServiceNameKey.String("otel-round-tripper"),
			),
		),
	}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10; i++ {
		// Add a random sleep duration so you will see the metrics in the console
		url := "https://httpstat.us/200?sleep=" + strconv.Itoa(rand.Intn(1000)+1000) //nolint:gosec

		log.Printf("GET: %s", url)
		response, err := client.Get(url)
		if err != nil {
			log.Panicf("cannot perform http request: %v", err)
		}

		_ = response.Body.Close()
	}
}
