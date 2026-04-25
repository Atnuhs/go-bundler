package lib

import "fmt"

const LibConst = 42

var (
	LibVar = "lib var"
)

type LibStruct struct {
	Value int
}

type LibInterface interface {
	Greet()
}

func LibFunc() {
	fmt.Println("from lib")
}
