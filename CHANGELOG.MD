# Changelog

## Unreleased

## v0.0.16

### Changes
- [x] Updated OpenTelemetry libraries to v1.47.0.
- [x] Updated the minimum Go version to 1.26.0, as required by OpenTelemetry.
- [x] Fixed CI/CD toolchain compatibility by using the Go version from `go.mod` for validation and tests.
- [x] Removed the Go Report Card badge from the README.

## Historical changes

### Changed
- There were breaking changes in the [opentelemetry-go Release v1.6.0/v0.28.0](https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.6.0) for the metric pkg. These also effect this package.
  - One of the changes is that `metric.Meter` is now an interface instead of a struct. It can be provided via `global.MeterProvider().Meter("name")` (see [example](example_test.go#L55))
