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

### Initializing the Client

An instance of the client can be created using `New()`.

```go
package main

import (
	"github.com/NdoleStudio/go-otelroundtripper"
)

func main()  {
	statusClient := client.New(client.WithDelay(200))
}
```

## Testing

You can run the unit tests for this client from the root directory using the command below:

```bash
go test -v
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
