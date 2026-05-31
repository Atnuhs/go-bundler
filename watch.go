package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounceInterval = 200 * time.Millisecond

func runWatch() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	outAbs, err := filepath.Abs(out)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	watched := make(map[string]bool)
	runBuild := func() {
		clearScreen()
		start := time.Now()
		fmt.Println(colorize("=== go-bundler watch rebuild ===", ansiBlue))
		dirs, err := safeRunOnce()
		if err != nil {
			fmt.Println(colorize("BUILD FAILED", ansiRed))
			fmt.Printf("%v\n\n", err)
			fmt.Printf("output unchanged: %s\n", out)
			fmt.Printf("watching %d directories\n", len(watched))
			return
		}
		reconcileWatches(watcher, watched, dirs)
		fmt.Println(colorize("BUILD SUCCEEDED", ansiGreen))
		fmt.Printf("rebuilt %s in %s; watching %d directories\n",
			out, time.Since(start).Round(time.Millisecond), len(watched))
	}

	// Builds run serially on a dedicated goroutine so reconcileWatches
	// never mutates `watched` concurrently. trigger is a coalescing
	// channel: while a build is in flight, additional requests collapse
	// into at most one pending build.
	trigger := make(chan struct{}, 1)
	buildDone := make(chan struct{})
	go func() {
		defer close(buildDone)
		for {
			select {
			case <-ctx.Done():
				return
			case <-trigger:
				runBuild()
			}
		}
	}()
	requestBuild := func() {
		select {
		case trigger <- struct{}{}:
		default:
		}
	}

	log.Print("starting initial build; press Ctrl+C to stop")
	requestBuild()

	// timer is only touched from this goroutine (schedule + cleanup),
	// so no mutex is required.
	var timer *time.Timer
	schedule := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceInterval, requestBuild)
	}

	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			<-buildDone
			log.Println("stopping watcher")
			return nil
		case ev, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if !shouldHandle(ev, outAbs) {
				continue
			}
			schedule()
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("watch error: %v", err)
		}
	}
}

// safeRunOnce wraps runOnce with a recover so a panic in the bundler
// (typically caused by syntactically invalid source during editing) does
// not kill the watcher.
func safeRunOnce() (dirs []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("bundler panicked (likely invalid Go in source): %v", r)
		}
	}()
	return runOnce()
}

const (
	ansiClear = "\033[H\033[2J"
	ansiRed   = "\033[31m"
	ansiGreen = "\033[32m"
	ansiBlue  = "\033[34m"
	ansiReset = "\033[0m"
)

func clearScreen() {
	fmt.Print(ansiClear)
}

func colorize(text, color string) string {
	return color + text + ansiReset
}

func shouldHandle(ev fsnotify.Event, outAbs string) bool {
	if filepath.Ext(ev.Name) != ".go" {
		return false
	}
	// CHMOD fires spuriously on some editors; we only care about content moves
	if ev.Op == fsnotify.Chmod {
		return false
	}
	if abs, err := filepath.Abs(ev.Name); err == nil && abs == outAbs {
		return false
	}
	return true
}

func reconcileWatches(w *fsnotify.Watcher, current map[string]bool, want []string) {
	wantSet := make(map[string]bool, len(want))
	for _, d := range want {
		wantSet[d] = true
		if !current[d] {
			if err := w.Add(d); err != nil {
				log.Printf("add watch %s: %v", d, err)
				continue
			}
			current[d] = true
		}
	}
	for d := range current {
		if !wantSet[d] {
			_ = w.Remove(d)
			delete(current, d)
		}
	}
}
