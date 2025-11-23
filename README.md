# go-bundler

go-bundler is a tool written in Go for merging source files for competitive programming contests like AtCoder.

## Install

```bash
go install github.com/Atnuhs/go-bundler@latest
```

## Usage

```bash
go-bundler -dir ./path/to/your/package > bundled.go
```

`go-bundler` bundles a Go package into a single source file. By default it only emits the bundled code.

You can enable additional comment blocks with the following flags:

```text
  -dir string
        target package directory (default ".")
  -with-metrics
        emit go-bundler metrics comment block
  -with-sustainability-metrics
        emit sustainability metrics (CO2, trees) in comment block
```

## Example

Emit a simple bundled file:

```bash
go-bundler -dir ./cmd/app > bundled.go
```

Emit a bundled file with line-count metrics:

```bash
go-bundler -dir ./cmd/app -with-metrics > bundled.go
```

Emit a bundled file with line-count metrics and sustainability metrics:

```bash
go-bundler -dir ./cmd/app -with-metrics -with-sustainability-metrics > bundled.go
```

When `-with-sustainability-metrics` is enabled, `go-bundler` appends an additional metrics block that

includes a rough model-based estimate of CO2 reduction and an equivalent number of trees planted.

These values are purely illustrative and are not intended to represent actual environmental impact.

## License

MIT License
