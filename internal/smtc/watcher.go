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
// PowerShell process. One spawn at startup; outputs a line on song change or playback state change.
type Watcher struct {
	OnPause  func(posMs int64)
	OnPlay   func(posMs int64)
	OnSeek   func(posMs int64)
	onChange func(NowPlaying)
	onError  func(error)
	mu       sync.Mutex
	last     NowPlaying
}

// NewWatcher creates a watcher. onError is called when the helper process fails; pass nil to silence.
// Set OnPause and OnPlay after construction to handle playback state changes.
func NewWatcher(_ time.Duration, onChange func(NowPlaying), _ func(int64), onError func(error)) *Watcher {
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

// smtcScript is a PowerShell polling loop compatible with both PS 5.1 and PS 7.
// Outputs message types separated by the ASCII Unit Separator (0x1F):
//
//	SONG<US>Artist<US>Title  — on song title/artist change
//	PAUSE                    — on playback paused/stopped
//	PLAY                     — on playback resumed
const smtcScript = `
$ErrorActionPreference = 'SilentlyContinue'
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$null = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager,Windows.Media.Control,ContentType=WindowsRuntime]

$asTaskMethod = [System.WindowsRuntimeSystemExtensions].GetMethods() |
    Where-Object { $_.Name -eq 'AsTask' -and $_.GetParameters().Count -eq 1 -and $_.IsGenericMethodDefinition } |
    Select-Object -First 1

function WinRT-Await($asyncOp, [Type]$resultType) {
    $task = $asTaskMethod.MakeGenericMethod($resultType).Invoke($null, @($asyncOp))
    $task.Wait(-1) | Out-Null
    $task.Result
}

$mgrType   = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager]
$assembly  = $mgrType.Assembly
$propsType = $assembly.GetType('Windows.Media.Control.GlobalSystemMediaTransportControlsSessionMediaProperties')

$sep = [char]0x1F
$lastKey = ''
$lastPlaying = $true
$lastMs = 0

while ($true) {
    try {
        $mgr = WinRT-Await ($mgrType::RequestAsync()) $mgrType
        $s = $mgr.GetCurrentSession()
        if ($null -ne $s) {
            $playing = $false
            try { $playing = $s.GetPlaybackInfo().PlaybackStatus.ToString() -eq 'Playing' } catch {}

            $ms = 0
            try { $ms = [long]$s.GetTimelineProperties().Position.TotalMilliseconds } catch {}

            if ($playing -ne $lastPlaying) {
                if ($playing) {
                    "PLAY$sep$ms"
                    [Console]::Out.Flush()
                    # Position dropped significantly — likely a new song started
                    if (($lastMs - $ms) -gt 5000) { $lastKey = '' }
                } else {
                    "PAUSE$sep$ms"
                    [Console]::Out.Flush()
                }
                $lastPlaying = $playing
                $lastMs = $ms
            } elseif ($playing) {
                # Detect seek: position jumped more than 3s away from expected
                $expected = $lastMs + 1000
                $drift = [Math]::Abs($ms - $expected)
                if ($drift -gt 3000) {
                    "SEEK$sep$ms"
                    [Console]::Out.Flush()
                }
                $lastMs = $ms

                $p = WinRT-Await ($s.TryGetMediaPropertiesAsync()) $propsType
                if ($null -ne $p -and $p.Title -ne '') {
                    $key = "$($p.Artist)##SEP##$($p.Title)"
                    if ($key -ne $lastKey) {
                        "SONG$sep$($p.Artist)$sep$($p.Title)$sep$ms"
                        [Console]::Out.Flush()
                        $lastKey = $key
                    }
                }
            }
        } else {
            if ($lastPlaying) {
                "PAUSE"
                [Console]::Out.Flush()
                $lastPlaying = $false
            }
            $lastKey = ''
        }
    } catch {}
    Start-Sleep -Milliseconds 1000
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

		parts := strings.SplitN(line, fieldSep, 4)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "SONG":
			if len(parts) < 3 {
				continue
			}
			var posMs int64
			if len(parts) >= 4 {
				fmt.Sscan(parts[3], &posMs)
			}
			np := NowPlaying{
				Artist:     strings.TrimSpace(parts[1]),
				Title:      strings.TrimSpace(parts[2]),
				PositionMs: posMs,
			}
			w.mu.Lock()
			w.last = np
			w.mu.Unlock()
			w.onChange(np)

		case "PAUSE":
			var posMs int64
			if len(parts) >= 2 {
				fmt.Sscan(parts[1], &posMs)
			}
			fmt.Printf("[smtc] PAUSE: %dms\n", posMs)
			if w.OnPause != nil {
				w.OnPause(posMs)
			}

		case "PLAY":
			var posMs int64
			if len(parts) >= 2 {
				fmt.Sscan(parts[1], &posMs)
			}
			fmt.Printf("[smtc] PLAY: %dms\n", posMs)
			if w.OnPlay != nil {
				w.OnPlay(posMs)
			}

		case "SEEK":
			var posMs int64
			if len(parts) >= 2 {
				fmt.Sscan(parts[1], &posMs)
			}
			fmt.Printf("[smtc] SEEK: %dms\n", posMs)
			if w.OnSeek != nil {
				w.OnSeek(posMs)
			}

		}
	}

	cmd.Wait()
}
