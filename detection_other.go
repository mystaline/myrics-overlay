//go:build !windows

package main

import "context"

// runDetection starts the ACRCloud audio fingerprint loop on non-Windows platforms.
// SMTC is Windows-only; on Linux/macOS audio capture feeds ACRCloud for recognition.
func (a *App) runDetection(ctx context.Context) {
	go a.acrCloudFallbackLoop(ctx)
}
