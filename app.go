package main

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds the application state
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// TODO: Step 1 - Configure window properties
	// Set window to always stay on top of other windows
	runtime.WindowSetAlwaysOnTop(ctx, true)

	// Size: 800x120 pixels, centered on screen
	runtime.WindowSetSize(ctx, 800, 120)
	runtime.WindowCenter(ctx)

	// Goroutine to continuously detect music
	go a.detectionLoop()
}

// detectionLoop runs continuously to detect playing music
func (a *App) detectionLoop() {
	// Check for music every 15 seconds
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// TODO: Step 2 - Capture audio snippet

		// TODO: Step 3 - Recognize song using ACRCloud

		// TODO: Step 4 - Fetch lyrics for detected song

		// TODO: Step 5 - Parse LRC format lyrics

		// TODO: Step 6 - Start lyrics synchronization
	}
}

// startLyricsSync synchronizes lyrics with playback time
func (a *App) startLyricsSync(lyrics []LyricLine) {
	// TODO: Step 7 - Track playback time
}

// getCurrentLyrics finds the current and next lyrics based on elapsed time
func (a *App) getCurrentLyrics(lyrics []LyricLine, elapsed time.Duration) (string, string) {
	// Loop through lyrics and find the line matching current time
	return "Sample current line", "Sample next line"
}

// LyricLine represents a single line of lyrics with timestamp
type LyricLine struct {
	Timestamp time.Duration
	Text      string
}

// Exported methods (callable from frontend)

// SearchSong allows manual song search from the frontend
func (a *App) SearchSong(title string) error {
	// TODO: Implement manual search
	// Fetch lyrics for the given song title
	return nil
}

// ToggleDetection starts or stops automatic music detection
func (a *App) ToggleDetection(enabled bool) {
	// TODO: Implement detection toggle
	// Start or stop the detection loop
}
