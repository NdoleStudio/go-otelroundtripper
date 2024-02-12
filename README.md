# go-otelroundtripper

[![Build](https://github.com/NdoleStudio/go-otelroundtripper/actions/workflows/main.yml/badge.svg)](https://github.com/NdoleStudio/go-otelroundtripper/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/NdoleStudio/go-otelroundtripper/branch/main/graph/badge.svg)](https://codecov.io/gh/NdoleStudio/go-otelroundtripper)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/NdoleStudio/go-otelroundtripper/badges/quality-score.png?b=main)](https://scrutinizer-ci.com/g/NdoleStudio/go-otelroundtripper/?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/NdoleStudio/go-otelroundtripper)](https://goreportcard.com/report/github.com/NdoleStudio/go-otelroundtripper)
[![GitHub contributors](https://img.shields.io/github/contributors/NdoleStudio/go-otelroundtripper)](https://github.com/NdoleStudio/go-otelroundtripper/graphs/contributors)
[![GitHub license](https://img.shields.io/github/license/NdoleStudio/go-otelroundtripper?color=brightgreen)](https://github.com/NdoleStudio/go-otelroundtripper/blob/master/LICENSE)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/NdoleStudio/go-otelroundtripper)](https://pkg.go.dev/github.com/NdoleStudio/go-otelroundtripper)


This package provides an easy way to collect http related metrics
(e.g Response times, Status Codes, number of in flight requests etc) for your HTTP API Clients.
You can do this by using this round tripper when instantiating the `http.Client{}`.

## Why this package exists

I currently have to integrate with multiple APIs and I needed a simple way to export metrics for those external
API's. Sometimes external API's have their own SDK and the only input is `http.Client`. In this scenario, I can create
an HTTP client with a round tripper automatically exports metrics according to the open telemetry specification.

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

This is a sample code that instantiates an HTTP client which sends requests to `https://example.com`.
You can see a runnable [example here](./example_test.go)

```go
client := http.Client{
    Transport: New(
        WithName("example.com")
        WithMeter(global.MeterProvider()Meter("otel-round-tripper")),
        WithAttributes(
            semconv.ServiceNameKey.String("otel-round-tripper"),
        ),
    ),
}

response, err := client.Get("https://example.com")
```

## Metrics Emitted

The following metrics will be emitted by this package. Note that `*` will be replaced by the prefix passed in `WithName()`.

- `*.no_request` http calls with nil `http.Request`
- `*.errors` http requests which had an error response i.e `err != nil`
- `*.success` http requests which were successfull. Meaning there were no transport errors
- `*.timeouts` http requests which timed out
- `*.cancelled` http requests with cancelled context
- `*.deadline_exceeded` http requests with context dateline exceeded
- `*.total_duration` total time it takes to execute the http request in milliseconds
- `*.in_flight` concurrent http requests
- `*.attempts` http requests attempted
- `*.failures` http requests with  http status code >= 400
- `*.redirects`  http requests with  300 <= http status code < 400

## Testing

You can run the unit tests for this client from the root directory using the command below:

```bash
go test -v
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
