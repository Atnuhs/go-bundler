package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

var (
	withMetrics               = flag.Bool("with-metrics", false, "emit go-bundler metrics comment block")
	withSustainabilityMetrics = flag.Bool("with-sustainability-metrics", false, "emit sustainability metrics (CO2, trees) in comment block")
	dir                       = flag.String("dir", ".", "target package directory")
	watchMode                 = flag.Bool("watch", false, "watch local package files and rebuild on change (requires -o)")
	out                       string
)

func init() {
	flag.StringVar(&out, "out", "", "output file path (default: stdout)")
	flag.StringVar(&out, "o", "", "output file path (shorthand)")
}

func main() {
	flag.Parse()

	if *watchMode {
		if out == "" {
			log.Fatal("-watch requires -o/--out to be set")
		}
		if err := runWatch(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if _, err := runOnce(); err != nil {
		log.Fatal(err)
	}
}

// runOnce performs one full bundle pass and returns the local module
// package directories observed during load (for the watcher to monitor).
func runOnce() ([]string, error) {
	pkgs, err := loadPackages(*dir)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	var raw bytes.Buffer
	originalLines, err := Bundle(pkgs, &raw)
	if err != nil {
		return nil, fmt.Errorf("bundle: %w", err)
	}

	formatted, err := imports.Process("main.go", raw.Bytes(), &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  4,
	})
	if err != nil {
		return nil, fmt.Errorf("goimports: %w", err)
	}
	bundledLines := bytes.Count(formatted, []byte{'\n'})

	var w io.Writer = os.Stdout
	if out != "" {
		f, err := os.Create(out)
		if err != nil {
			return nil, fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()
		w = f
	}
	g := NewMetricWriter(originalLines, bundledLines)
	g.WriteHeader(w)
	if *withMetrics {
		g.WriteMetrics(w)
	}
	if *withSustainabilityMetrics {
		g.WriteSustainabilityMetrics(w)
	}
	g.WriteProjectURL(w)
	if _, err := w.Write(formatted); err != nil {
		return nil, fmt.Errorf("write output: %w", err)
	}

	return localPackageDirs(pkgs), nil
}

func loadPackages(dir string) ([]*packages.Package, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get abs path of %s", dir)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedDeps |
			packages.NeedModule |
			packages.NeedCompiledGoFiles |
			packages.NeedImports,
		Dir:   absDir,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	return pkgs, nil
}
