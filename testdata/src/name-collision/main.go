package main

import (
	liba "github.com/Atnuhs/go-bundler/testdata/src/name-collision/liba"
	libb "github.com/Atnuhs/go-bundler/testdata/src/name-collision/libb"
)

func main() {
	liba.FuncA()
	libb.FuncB()
}
