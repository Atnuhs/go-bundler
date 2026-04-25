package main

import (
	. "github.com/Atnuhs/go-bundler/testdata/src/dot-import/lib"
)

func useLib(x LibStruct) LibStruct { return x }

func main() {
	defer LibFunc()
	s := LibStruct{Value: 42}
	_ = useLib(s)
	_ = LibVar
	_ = LibConst
	var _ LibInterface = nil
	_ = any(s).(LibInterface)
}
