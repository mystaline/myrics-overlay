//go:build !windows

package smtc

import (
	"context"
	"time"
)

// NowPlaying holds current media info.
type NowPlaying struct {
	Title      string
	Artist     string
	PositionMs int64
}

// Watcher is a no-op on non-Windows platforms.
// On Linux/macOS, detection is handled by the ACRCloud fallback loop.
type Watcher struct{}

func NewWatcher(_ time.Duration, _ func(NowPlaying), _ func(error)) *Watcher {
	return &Watcher{}
}

func (w *Watcher) Start(_ context.Context) {}
