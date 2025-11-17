package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestMain(m *testing.M) {
	Level.Set(slog.LevelDebug)
	os.Exit(m.Run())
}

func buildPackages(t *testing.T, dir string) []*packages.Package {
	t.Helper()
	absDir, _ := filepath.Abs(filepath.Join("testdata/src", dir))
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
		t.Fatal(err)
	}
	return pkgs
}

func TestBundler(t *testing.T) {
	tests := []struct {
		name    string
		testdir string
		wantErr bool
	}{
		{
			name:    "no dependencies",
			testdir: "no-deps",
		},
		{
			name:    "single dependencies",
			testdir: "single-deps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare
			pkgs := buildPackages(t, tt.testdir)

			// execute
			result, err := Bundle(pkgs)

			// validate
			if (err != nil) != tt.wantErr {
				t.Errorf("Bundle() error = %v, wantErr %v", err, tt.wantErr)
			}
			// 結果の検証...
			t.Log(result)
		})
	}
}
