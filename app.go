package main

import (
	"context"
	"sync"
	"time"

	"github.com/mystaline/myrics-overlay/internal/config"
	"github.com/mystaline/myrics-overlay/internal/lyrics"
	"github.com/mystaline/myrics-overlay/internal/recognition"
	"github.com/mystaline/myrics-overlay/pkg/models"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application struct.
type App struct {
	ctx               context.Context
	cfg               *config.Config
	recognitionClient *recognition.Client
	lyricsFetcher     *lyrics.Fetcher

	mu               sync.Mutex
	currentSong      string // "Artist|Title" key to detect song changes
	detectionEnabled bool
	detectionCancel  context.CancelFunc
}

// NewApp creates the App, wiring up all clients from config.
func NewApp(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
		recognitionClient: recognition.NewClient(
			cfg.ACRCloud.AccessKey,
			cfg.ACRCloud.SecretKey,
			cfg.ACRCloud.Host,
		),
		lyricsFetcher: lyrics.NewFetcher(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	runtime.WindowSetSize(ctx, 800, 100)
	runtime.WindowSetPosition(ctx, 500, 0)
	runtime.WindowShow(ctx)
	runtime.WindowSetAlwaysOnTop(ctx, true)

	a.detectionEnabled = true
	go a.startDetection()
}

// startDetection sets up a cancellable context and starts the platform-specific
// detection mechanism. Guards against double-start from rapid ToggleDetection calls.
func (a *App) startDetection() {
	detCtx, cancel := context.WithCancel(a.ctx)

	a.mu.Lock()
	if a.detectionCancel != nil {
		// Already running — a second concurrent startDetection call lost the race.
		a.mu.Unlock()
		cancel()
		return
	}
	a.detectionCancel = cancel
	a.mu.Unlock()

	a.runDetection(detCtx)
}

// onSongChanged is called when detection (SMTC or ACRCloud) reports a new song.
// Fetches and emits lyrics to the frontend.
func (a *App) onSongChanged(title, artist string, offsetMs int64) {
	if title == "" {
		return
	}

	songKey := artist + "|" + title
	a.mu.Lock()
	if songKey == a.currentSong {
		a.mu.Unlock()
		return
	}
	a.currentSong = songKey
	a.mu.Unlock()

	runtime.LogInfof(a.ctx, "Now playing: %s - %s (offset: %dms)", artist, title, offsetMs)
	runtime.EventsEmit(a.ctx, "now-playing", map[string]string{
		"title":  title,
		"artist": artist,
	})

	detectedAt := time.Now()

	lyricsContent, err := a.lyricsFetcher.FetchLyrics(&models.SongInfo{
		Title:  title,
		Artist: artist,
	})
	if err != nil {
		runtime.LogWarningf(a.ctx, "Lyrics fetch failed: %v", err)
		runtime.EventsEmit(a.ctx, "lyrics-error", "No lyrics found for this song")
		// Clear so the next SMTC poll can retry for the same song.
		a.mu.Lock()
		a.currentSong = ""
		a.mu.Unlock()
		return
	}

	// Compensate for the time spent fetching so sync starts at the right position.
	adjustedOffsetMs := offsetMs + time.Since(detectedAt).Milliseconds()

	parsedLyrics, err := lyrics.ParseLRC(lyricsContent)
	if err != nil {
		// LRCLIB returned plain (unsynced) lyrics — display as static text.
		runtime.EventsEmit(a.ctx, "lyrics-plain", lyricsContent)
		return
	}

	runtime.LogInfof(a.ctx, "Lyrics loaded: %d lines", len(parsedLyrics))
	runtime.EventsEmit(a.ctx, "lyrics-found", map[string]any{
		"lines":    parsedLyrics,
		"offsetMs": adjustedOffsetMs,
	})
}

// SearchSong triggers a manual lyrics lookup from the frontend.
func (a *App) SearchSong(title, artist string) {
	a.onSongChanged(title, artist, 0)
}

// TriggerTest triggers a test song without audio capture (dev/debug helper).
func (a *App) TriggerTest(title, artist string) {
	runtime.LogInfof(a.ctx, "Test trigger: %s - %s", artist, title)
	a.mu.Lock()
	a.currentSong = ""
	a.mu.Unlock()
	a.onSongChanged(title, artist, 0)
}

// ToggleDetection pauses or resumes detection from the frontend.
func (a *App) ToggleDetection(enabled bool) {
	a.mu.Lock()
	same := enabled == a.detectionEnabled
	a.detectionEnabled = enabled
	if !enabled && a.detectionCancel != nil {
		a.detectionCancel()
		a.detectionCancel = nil
	}
	a.mu.Unlock()

	if same {
		return
	}
	if enabled {
		go a.startDetection()
	}
}
