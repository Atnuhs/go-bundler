package main

import "fmt"
// github.com/Atnuhs/go-bundler/testdata/src/no-deps/main.go:5:1
func main() {
	main_inner()
}
// github.com/Atnuhs/go-bundler/testdata/src/no-deps/main.go:9:1
func main_inner() {
	fmt.Println("hoge")
}
