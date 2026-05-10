//go:build windows

package main

import (
	"context"
	"time"

	"github.com/mystaline/myrics-overlay/internal/smtc"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// runDetection starts the Windows SMTC watcher.
// SMTC polls Windows "Now Playing" via PowerShell every 3 seconds,
// calling onSongChanged whenever the song title/artist changes.
func (a *App) runDetection(ctx context.Context) {
	watcher := smtc.NewWatcher(
		3*time.Second,
		func(np smtc.NowPlaying) { a.onSongChanged(np.Title, np.Artist, np.PositionMs) },
		func(err error) { runtime.LogWarningf(a.ctx, "SMTC error: %v", err) },
	)
	watcher.Start(ctx)
}
