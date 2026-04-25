package main

import (
	. "github.com/Atnuhs/go-bundler/testdata/src/dot-import/lib"
)

func main() {
	LibFunc()
	s := LibStruct{Value: 42}
	_ = s
}
