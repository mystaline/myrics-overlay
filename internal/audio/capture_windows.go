//go:build windows

package audio

import (
	"fmt"
	"time"
)

// Capturer is a no-op on Windows.
// SMTC (System Media Transport Controls) is the primary detection mechanism on Windows,
// so audio capture is not needed. The ACRCloud fallback loop will not start if
// NewCapturer returns an error.
type Capturer struct{}

func NewCapturer() (*Capturer, error) {
	return nil, fmt.Errorf("audio capture not supported on Windows; SMTC handles detection")
}

func (c *Capturer) CaptureSnippet(_ time.Duration) ([]byte, error) {
	return nil, fmt.Errorf("audio capture not supported on Windows")
}

func (c *Capturer) Close() error { return nil }
