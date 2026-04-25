---
title: go-bundler - Go source bundler for AtCoder
description: A lightweight Go source bundler for competitive programming and AtCoder single-file submissions.
---

# go-bundler

**go-bundler** is a lightweight Go source bundler for competitive programming.

It merges multiple Go source files into a single Go file, making it easier to
submit Go programs to AtCoder, yukicoder, and other online judges that require
single-file submissions.

## Features

- Bundle a Go package into a single source file
- Dead code elimination via RTA (Rapid Type Analysis)
- Supports generics, embedded structs, and interface types
- Single-command usage, outputs to stdout
- Optional line-count and sustainability metrics

## Install

```sh
go install github.com/Atnuhs/go-bundler@latest
```

## Usage

```sh
go-bundler -dir ./path/to/your/package > submit.go
```

### Options

| Flag | Description |
|---|---|
| `-dir` | Target package directory (default: `.`) |
| `-with-metrics` | Emit line-count metrics as a comment block |
| `-with-sustainability-metrics` | Emit CO2 and tree-equivalent metrics |

## Examples

Bundle a package and write to a file:

```sh
go-bundler -dir ./cmd/abc123 > submit.go
```

Bundle with line-count metrics:

```sh
go-bundler -dir ./cmd/abc123 -with-metrics > submit.go
```

## Links

- [GitHub repository](https://github.com/Atnuhs/go-bundler)
- [Qiita: Go のバンドラーを書いた](https://qiita.com/Authns/items/ddba6d392ec6a316383f)
