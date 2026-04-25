package main

import "fmt"

func init() {

	main_init()

	lib_init()
}
func main_init() {
	fmt.Println("hoge")
	main_init_sub()
}

func lib_init() {
	lib_Foo1 = 10
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:33:6
type main_Embedded struct {
	lib_LibStruct
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:41:6
type main_NonEmbedded struct {
	s lib_LibStruct
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:49:6
type main_Seeker interface {
	Seek()
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:5:6
type lib_V int
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:7:6
type lib_LibStruct struct {
	V lib_V
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:40:6
type lib_Seeker[T any] struct {
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:15:5
var lib_LibStruct1 = lib_LibStruct{}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:16:5
var lib_Foo1 = 0
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:18:1
const (
	main_X1 = iota
	main_X2
	main_X3
)
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:29:1
const main_HOGE11, main_HOGE12 = lib_HOGE1, lib_HOGE2
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:23:1
const (
	lib_HOGE1 = 1
	lib_HOGE2 = 1
)
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:68:1
func main() {
	lib_LibFunc()
	lib_LibStruct1.V = 10
	data := main_Embedded{lib_LibStruct{}}
	data2 := main_Embedded{lib_LibStruct: lib_LibStruct{}}
	data3 := main_NonEmbedded{s: lib_LibStruct{}}
	fmt.Println(data.V)
	fmt.Println(data2.V)
	fmt.Println(data3.s.V)
	fmt.Println(main_HOGE11)
	fmt.Println(main_X1)
	main_FunctionWithArg(10)
	main_SeekerSeek(lib_NewSeeker[int]())

}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:9:1
func main_init_sub() {

}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:37:1
func (d main_Embedded) String() {
	fmt.Println(d.lib_LibStruct.V)
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:45:1
func (d main_NonEmbedded) String() {
	fmt.Println(d.s.V)
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:59:1
func main_SeekerSeek(s main_Seeker) {
	s.Seek()
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/main.go:63:1
func main_FunctionWithArg(x int) {
	var inner = 1
	fmt.Println(x, inner)
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:11:1
func (v lib_LibStruct) Print() {
	fmt.Println(v.V)
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:36:1
func lib_LibFunc() {
	fmt.Println("from lib")
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:43:1
func lib_NewSeeker[T any]() lib_Seeker[T] {
	return lib_Seeker[T]{}
}
// github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib/lib.go:47:1
func (s lib_Seeker[T]) Seek() {
	fmt.Println("seeker is seeking")
}
