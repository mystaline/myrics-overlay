//go:build windows

package smtc

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// NowPlaying holds current media info from Windows SMTC.
type NowPlaying struct {
	Title      string
	Artist     string
	PositionMs int64
}

// fieldSep is ASCII Unit Separator (0x1F) — cannot appear in media metadata.
const fieldSep = "\x1f"

// Watcher reads Windows System Media Transport Controls via a single long-running
// PowerShell process. One spawn at startup; outputs a line only when the song changes.
type Watcher struct {
	interval time.Duration // kept for API compat; PS script handles its own sleep
	onChange func(NowPlaying)
	onError  func(error)
	mu       sync.Mutex
	last     NowPlaying
}

// NewWatcher creates a watcher. onError is called when the helper process fails
// to start or exits unexpectedly; pass nil to silence it.
func NewWatcher(_ time.Duration, onChange func(NowPlaying), onError func(error)) *Watcher {
	return &Watcher{onChange: onChange, onError: onError}
}

// Start launches the long-running SMTC helper process; restarts it if it exits
// unexpectedly. Stops permanently when ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) {
	go func() {
		for {
			w.run(ctx)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				// brief pause before restarting after unexpected exit
			}
		}
	}()
}

// smtcScript is a PowerShell polling loop that runs inside a single persistent process.
// It outputs "Artist<US>Title<US>PositionMs" only when the song title/artist changes,
// using the ASCII Unit Separator (0x1F) to avoid conflicts with artist/title content.
const smtcScript = `
$ErrorActionPreference = 'SilentlyContinue'
$null = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager,Windows.Media.Control,ContentType=WindowsRuntime]
$sep = [char]0x1F
$lastKey = ''

while ($true) {
    try {
        $mgr = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager]::RequestAsync().GetAwaiter().GetResult()
        $s = $mgr.GetCurrentSession()
        if ($null -ne $s) {
            $p = $s.TryGetMediaPropertiesAsync().GetAwaiter().GetResult()
            if ($null -ne $p -and $p.Title -ne '') {
                $key = "$($p.Artist)##SEP##$($p.Title)"
                if ($key -ne $lastKey) {
                    $t = $s.GetTimelineProperties()
                    $ms = [long]$t.Position.TotalMilliseconds
                    "$($p.Artist)$sep$($p.Title)$sep$ms"
                    [Console]::Out.Flush()
                    $lastKey = $key
                }
            }
        } else {
            $lastKey = ''
        }
    } catch {}
    Start-Sleep -Milliseconds 3000
}
`

func (w *Watcher) run(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "powershell",
		"-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden",
		"-Command", smtcScript)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		if w.onError != nil {
			w.onError(fmt.Errorf("SMTC pipe: %w", err))
		}
		return
	}

	if err := cmd.Start(); err != nil {
		if w.onError != nil {
			w.onError(fmt.Errorf("SMTC start: %w", err))
		}
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, fieldSep, 3)
		if len(parts) < 2 {
			continue
		}

		np := NowPlaying{
			Artist: strings.TrimSpace(parts[0]),
			Title:  strings.TrimSpace(parts[1]),
		}
		if len(parts) == 3 {
			if _, err := fmt.Sscan(parts[2], &np.PositionMs); err != nil {
				np.PositionMs = 0
			}
		}

		// PS script only outputs on song change, so every line is a new song.
		// Reset position — stale WinRT timeline data can appear at track boundaries.
		np.PositionMs = 0

		w.mu.Lock()
		w.last = np
		w.mu.Unlock()

		w.onChange(np)
	}

	cmd.Wait()
}
