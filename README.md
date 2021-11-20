# go-otelroundtripper

[![Build](https://github.com/NdoleStudio/go-otelroundtripper/actions/workflows/main.yml/badge.svg)](https://github.com/NdoleStudio/go-otelroundtripper/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/NdoleStudio/go-otelroundtripper/branch/main/graph/badge.svg)](https://codecov.io/gh/NdoleStudio/go-otelroundtripper)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/NdoleStudio/go-otelroundtripper/badges/quality-score.png?b=main)](https://scrutinizer-ci.com/g/NdoleStudio/go-otelroundtripper/?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/NdoleStudio/go-otelroundtripper)](https://goreportcard.com/report/github.com/NdoleStudio/go-otelroundtripper)
[![GitHub contributors](https://img.shields.io/github/contributors/NdoleStudio/go-otelroundtripper)](https://github.com/NdoleStudio/go-otelroundtripper/graphs/contributors)
[![GitHub license](https://img.shields.io/github/license/NdoleStudio/go-otelroundtripper?color=brightgreen)](https://github.com/NdoleStudio/go-otelroundtripper/blob/master/LICENSE)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/NdoleStudio/go-otelroundtripper)](https://pkg.go.dev/github.com/NdoleStudio/go-otelroundtripper)


This package provides an easy way to collect http related metrics(e.g Response times, Status Codes, Number of inflight requests etc) for your HTTP API Clients.
You can do this by passing in this round tripper when instantiating the `http.CLient{}`.

## Installation

`go-otelroundtripper` is compatible with modern Go releases in module mode, with Go installed:

```bash
go get github.com/NdoleStudio/go-otelroundtripper
```

Alternatively the same can be achieved if you use `import` in a package:

```go
import "github.com/NdoleStudio/go-otelroundtripper"
```

## Usage

### Using the RoundTripper

This is a sample application that instantiates an http client which sends requests to `https://httpstat.us`.
The open telemetry metrics will be exported to stdout.

```go
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


func main() {
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

	// Here we are using the http client created above perform 10 http requests to
	// https://httpstat.us/200. The metrics will be exported to the console.
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10; i++ {
		// Add a random sleep duration so as not to DDOS the https://httstat.us website
		url := "https://httpstat.us/200?sleep=" + strconv.Itoa(rand.Intn(1000) + 1000)

		log.Printf("GET: %s", url)
		response, err := client.Get(url)
		if err != nil {
			log.Panicf("cannot perform http request: %v", err)
		}

		_ = response.Body.Close()
	}
}
```

## Testing

You can run the unit tests for this client from the root directory using the command below:

```bash
go test -v
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
