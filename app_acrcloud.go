//go:build !windows

package main

import (
	"context"
	"time"

	"github.com/mystaline/myrics-overlay/internal/audio"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// acrCloudFallbackLoop captures audio and fingerprints it via ACRCloud.
// This is the primary detection path on Linux/macOS where SMTC is unavailable.
func (a *App) acrCloudFallbackLoop(ctx context.Context) {
	capturer, err := audio.NewCapturer()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Audio capturer init failed: %v", err)
		return
	}
	defer capturer.Close()

	interval := max(
		time.Duration(a.cfg.Detection.Interval)*time.Second,
		15*time.Second,
	)

	for {
		a.runACRCloudDetection(capturer)
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}

func (a *App) runACRCloudDetection(capturer *audio.Capturer) {
	audioData, err := capturer.CaptureSnippet(10 * time.Second)
	if err != nil {
		runtime.LogWarningf(a.ctx, "Audio capture failed: %v", err)
		return
	}

	songInfo, err := a.recognitionClient.Recognize(audioData)
	if err != nil {
		runtime.LogWarningf(a.ctx, "ACRCloud recognition failed: %v", err)
		return
	}
	if songInfo == nil {
		// No music detected — expected when nothing is playing.
		return
	}

	a.onSongChanged(songInfo.Title, songInfo.Artist, 0)
}
