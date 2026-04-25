package main

import (
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
	var buf strings.Builder
	if _, err := Bundle(pkgs, &buf); err != nil {
		t.Fatalf("Bundle() error = %v", err)
	}
	return buf.String()
}

func assertContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("expected %q in output\ngot:\n%s", substr, output)
	}
}

func assertNotContains(t *testing.T, output, substr string) {
	t.Helper()
	if strings.Contains(output, substr) {
		t.Errorf("unexpected %q in output\ngot:\n%s", substr, output)
	}
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
			output := bundleDir(t, tt.testdir)

			goldenPath := filepath.Join("testdata/golden", tt.testdir+".go")
			if *update {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(output), 0644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden (run with -update to generate): %v", err)
			}
			if output != string(want) {
				t.Errorf("Bundle() output mismatch\ngot:\n%s\nwant:\n%s", output, string(want))
			}
		})
	}
}

func TestTreeShaking(t *testing.T) {
	output := bundleDir(t, "tree-shaking")

	assertContains(t, output, "lib_UsedFunc")
	assertNotContains(t, output, "lib_UnusedFunc")
}

func TestDotImport(t *testing.T) {
	output := bundleDir(t, "dot-import")

	assertContains(t, output, "lib_LibFunc")
	assertContains(t, output, "lib_LibStruct")
	assertNotContains(t, output, ". \"")
}

func TestNameCollision(t *testing.T) {
	output := bundleDir(t, "name-collision")

	// 同名パッケージ(package lib)が2つある場合、パスのアルファベット順に lib_00, lib_01 が割り当てられる
	// liba → lib_00, libb → lib_01
	assertContains(t, output, "lib_00_FuncA")
	assertContains(t, output, "lib_01_FuncB")
}
