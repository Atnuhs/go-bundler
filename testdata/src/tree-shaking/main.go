package main

import "github.com/Atnuhs/go-bundler/testdata/src/tree-shaking/lib"

func main() {
	lib.UsedFunc()
}
