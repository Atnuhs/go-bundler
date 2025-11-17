package main

import (
	"log/slog"
	"os"
)

var Level = new(slog.LevelVar)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: Level,
	})))
}

func main() {
	println("Hello world")
}
