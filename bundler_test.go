package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

var update = flag.Bool("update", false, "update golden files")

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func loadTestPackage(t *testing.T, dir string) []*packages.Package {
	t.Helper()
	pkgs, err := loadPackages(filepath.Join("testdata/src", dir))
	if err != nil {
		t.Fatal(err)
	}
	return pkgs
}

func bundleDir(t *testing.T, dir string) string {
	t.Helper()
	pkgs := loadTestPackage(t, dir)
	var buf bytes.Buffer
	if _, err := Bundle(pkgs, &buf); err != nil {
		t.Fatalf("Bundle() error = %v", err)
	}
	return buf.String()
}

func TestGolden(t *testing.T) {
	tests := []struct {
		name    string
		testdir string
	}{
		{name: "no dependencies", testdir: "no-deps"},
		{name: "single dependencies", testdir: "single-deps"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := loadTestPackage(t, tt.testdir)
			var buf bytes.Buffer
			if _, err := Bundle(pkgs, &buf); err != nil {
				t.Fatalf("Bundle() error = %v", err)
			}

			goldenPath := filepath.Join("testdata/golden", tt.testdir+".go")
			if *update {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(goldenPath, buf.Bytes(), 0644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden (run with -update to generate): %v", err)
			}
			if got := buf.String(); got != string(want) {
				t.Errorf("Bundle() output mismatch\ngot:\n%s\nwant:\n%s", got, string(want))
			}
		})
	}
}

func TestTreeShaking(t *testing.T) {
	output := bundleDir(t, "tree-shaking")

	if !strings.Contains(output, "lib_UsedFunc") {
		t.Error("reachable function lib_UsedFunc should be present in output")
	}
	if strings.Contains(output, "lib_UnusedFunc") {
		t.Error("unreachable function lib_UnusedFunc should be eliminated")
	}
}

func TestNameCollision(t *testing.T) {
	output := bundleDir(t, "name-collision")

	// 同名パッケージ(package lib)が2つある場合、パスのアルファベット順に lib_00, lib_01 が割り当てられる
	// liba → lib_00, libb → lib_01
	if !strings.Contains(output, "lib_00_FuncA") {
		t.Errorf("expected lib_00_FuncA in output, got:\n%s", output)
	}
	if !strings.Contains(output, "lib_01_FuncB") {
		t.Errorf("expected lib_01_FuncB in output, got:\n%s", output)
	}
}
