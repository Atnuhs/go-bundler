package main

import (
	"golang.org/x/tools/go/packages"
)

var stdSet = buildStdSet()

func buildStdSet() map[pkgPath]bool {
	cfg := &packages.Config{
		Mode: packages.NeedName,
	}
	pkgs, err := packages.Load(cfg, "std")
	if err != nil {
		panic("std packages load failed")
	}
	std := make(map[pkgPath]bool, len(pkgs))
	for _, p := range pkgs {
		std[pkgPath(p.PkgPath)] = true
	}
	return std
}

func isStd(pp pkgPath) bool {
	return stdSet[pp]
}
