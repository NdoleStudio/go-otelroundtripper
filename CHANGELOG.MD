# Changelog

## Unreleased

### Changed
- There were breaking changes in the [opentelemetry-go Release v1.6.0/v0.28.0](https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.6.0) for the metric pkg. These also effect this package.
  - One of the changes is that `metric.Meter` is now an interface instead of a struct. It can be provided via `global.MeterProvider().Meter("name")` (see [example](example_test.go#L55))
