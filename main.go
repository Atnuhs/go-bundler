package main

import (
	"bytes"
	"flag"
	"fmt"
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
)

func main() {
	flag.Parse()

	pkgs, err := loadPackages(*dir)
	if err != nil {
		log.Fatalf("load packages: %v", err)
	}

	// execute summarize
	var raw bytes.Buffer
	originalLines, err := Bundle(pkgs, &raw)
	if err != nil {
		log.Fatalf("bundle: %v", err)
	}

	// format with goimports
	formatted, err := imports.Process("main.go", raw.Bytes(), &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  4,
	})
	if err != nil {
		log.Fatalf("goimports: %v", err)
	}
	bundledLines := bytes.Count(formatted, []byte{'\n'})

	// output
	w := os.Stdout
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
		log.Fatalf("write stdout: %v", err)
	}
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
