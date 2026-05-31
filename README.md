# go-bundler

🇯🇵 日本語版: [README.ja.md](README.ja.md)

Documentation: <https://atnuhs.github.io/go-bundler/>  
Article: <https://qiita.com/Authns/items/ddba6d392ec6a316383f>

go-bundler is a lightweight Go source bundler for competitive programming.
It recursively resolves local Go package files, removes unreachable code (tree shaking),
and emits a single bundled file — mainly for AtCoder, yukicoder, and other online judges
where single-file submission is required.

## Features

- Dead code elimination via RTA (Rapid Type Analysis)
- Supports generics, embedded structs, and interface types
- Single-command usage, outputs to stdout or a file
- Watch mode that rebuilds on file changes
- Optional line-count and sustainability metrics

## Install

```bash
go install github.com/Atnuhs/go-bundler@latest
```

## Usage

```bash
go-bundler -dir ./path/to/your/package > bundled.go
```

`go-bundler` bundles a Go package into a single source file. By default it only emits the bundled code to stdout.

You can write directly to a file with `-o`/`--out`, and rebuild automatically on changes with `-watch`:

```text
  -dir string
        target package directory (default ".")
  -o string
        output file path (shorthand for -out)
  -out string
        output file path (default: stdout)
  -watch
        watch local package files and rebuild on change (requires -o)
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

Write directly to a file instead of redirecting stdout:

```bash
go-bundler -dir ./cmd/app -o bundled.go
```

Watch mode rebuilds `bundled.go` whenever any `.go` file in the resolved local packages changes (Ctrl+C to stop):

```bash
go-bundler -dir ./cmd/app -o bundled.go -watch
```

When `-with-sustainability-metrics` is enabled, `go-bundler` appends an additional metrics block that

includes a rough model-based estimate of CO2 reduction and an equivalent number of trees planted.

These values are purely illustrative and are not intended to represent actual environmental impact.

## License

MIT License
