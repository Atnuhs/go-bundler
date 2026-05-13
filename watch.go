package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/tools/go/packages"
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

	build := func() {
		start := time.Now()
		var (
			dirs []string
			err  error
		)
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("bundler panicked (likely invalid Go in source): %v", r)
				}
			}()
			dirs, err = runOnce()
		}()
		if err != nil {
			log.Printf("rebuild failed: %v", err)
			return
		}
		log.Printf("rebuilt %s in %s", out, time.Since(start).Round(time.Millisecond))
		reconcileWatches(watcher, watched, dirs)
	}

	// initial build
	log.Printf("watching: starting initial build")
	build()
	log.Printf("watching %d directories; press Ctrl+C to stop", len(watched))

	var (
		timerMu sync.Mutex
		timer   *time.Timer
	)
	schedule := func() {
		timerMu.Lock()
		defer timerMu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceInterval, build)
	}

	for {
		select {
		case <-ctx.Done():
			timerMu.Lock()
			if timer != nil {
				timer.Stop()
			}
			timerMu.Unlock()
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

func shouldHandle(ev fsnotify.Event, outAbs string) bool {
	if filepath.Ext(ev.Name) != ".go" {
		return false
	}
	// CHMOD fires spuriously on some editors; we only care about content moves
	if ev.Op == fsnotify.Chmod {
		return false
	}
	if outAbs != "" {
		if abs, err := filepath.Abs(ev.Name); err == nil && abs == outAbs {
			return false
		}
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

// localPackageDirs returns unique directories containing Go source files
// from packages belonging to the same module as the root package(s).
// Standard library and third-party module sources are excluded.
func localPackageDirs(pkgs []*packages.Package) []string {
	var rootModPath string
	for _, p := range pkgs {
		if p.Module != nil {
			rootModPath = p.Module.Path
			break
		}
	}
	if rootModPath == "" {
		return nil
	}

	seen := make(map[string]bool)
	var dirs []string
	packages.Visit(pkgs, func(p *packages.Package) bool {
		if p.Module == nil || p.Module.Path != rootModPath {
			return false
		}
		for _, f := range p.GoFiles {
			d := filepath.Dir(f)
			if !seen[d] {
				seen[d] = true
				dirs = append(dirs, d)
			}
		}
		return true
	}, nil)
	return dirs
}
