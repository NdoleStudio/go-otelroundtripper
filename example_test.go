package otelroundtripper

//import (
//	"context"
//	"encoding/json"
//	"go.opentelemetry.io/otel/sdk/metric"
//	"log"
//	"math/rand"
//	"net/http"
//	"os"
//	"strconv"
//	"time"
//
//	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
//	"go.opentelemetry.io/otel/metric/global"
//	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
//)
//
//func InstallExportPipeline(ctx context.Context) func() {
//	// Print with a JSON encoder that indents with two spaces.
//	enc := json.NewEncoder(os.Stdout)
//	enc.SetIndent("", "  ")
//	exporter, err := stdoutmetric.New(stdoutmetric.WithEncoder(enc))
//	if err != nil {
//		log.Fatalf("creating stdoutmetric exporter: %v", err)
//	}
//
//	// Register the exporter with an SDK via a periodic reader.
//	sdk := metric.NewMeterProvider(
//		metric.WithReader(metric.NewPeriodicReader(exporter)),
//	)
//
//	global.SetMeterProvider(sdk)
//
//	return func() {
//		if err := sdk.Shutdown(ctx); err != nil {
//			log.Fatalf("stopping sdk: %v", err)
//		}
//	}
//}
//
//func Example() {
//	ctx := context.Background()
//
//	// Registers a meter Provider globally.
//	cleanup := InstallExportPipeline(ctx)
//	defer cleanup()
//
//	client := http.Client{
//		Transport: New(
//			WithMeter(global.MeterProvider().Meter("otel-round-tripper")),
//			WithAttributes(
//				semconv.ServiceNameKey.String("otel-round-tripper"),
//			),
//		),
//	}
//
//	rand.Seed(time.Now().UnixNano())
//	for i := 0; i < 10; i++ {
//		// Add a random sleep duration so that we will see the metrics in the console
//		url := "https://httpstat.us/200?sleep=" + strconv.Itoa(rand.Intn(1000)+1000) //nolint:gosec
//
//		log.Printf("GET: %s", url)
//		response, err := client.Get(url)
//		if err != nil {
//			log.Panicf("cannot perform http request: %v", err)
//		}
//
//		_ = response.Body.Close()
//	}
//}
