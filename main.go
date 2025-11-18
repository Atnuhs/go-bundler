package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

var Level = new(slog.LevelVar)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: Level,
	})))
}

func main() {
	dir := flag.String("dir", ".", "target package directory")
	flag.Parse()

	pkgs, err := loadPackages(*dir)
	if err != nil {
		log.Fatalf("load packages: %v", err)
	}

	if err := Bundle(pkgs, os.Stdout); err != nil {
		log.Fatalf("bundle: %v", err)
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
