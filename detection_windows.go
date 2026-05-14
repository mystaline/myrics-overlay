//go:build windows

package main

import (
	"context"
	"fmt"
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
		nil,
		func(err error) { runtime.LogWarningf(a.ctx, "SMTC error: %v", err) },
	)
	watcher.OnPause = func(_ int64) { runtime.EventsEmit(a.ctx, "playback-paused") }
	watcher.OnPlay = func(posMs int64) {
		fmt.Printf("Play event received with position: %d\n", posMs)
		runtime.EventsEmit(a.ctx, "playback-resumed", posMs)
	}
	watcher.OnSeek = func(posMs int64) {
		fmt.Printf("Seek event received with position: %d\n", posMs)
		runtime.EventsEmit(a.ctx, "playback-seeked", posMs)
	}
	watcher.Start(ctx)
}
