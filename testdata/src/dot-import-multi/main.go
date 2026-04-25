package main

import (
	. "github.com/Atnuhs/go-bundler/testdata/src/dot-import-multi/liba"
	. "github.com/Atnuhs/go-bundler/testdata/src/dot-import-multi/libb"
)

func main() {
	a := FuncA()
	b := FuncB()
	_ = TypeA{Value: a.Value}
	_ = TypeB{Value: b.Value}
}
