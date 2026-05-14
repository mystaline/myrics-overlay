# Myrics

A lyrics overlay for Windows that syncs with whatever you're playing on YouTube Music. Built with Go + Wails.

## How it works

Myrics reads the currently playing song from Windows SMTC (System Media Transport Controls), fetches synced lyrics, and displays them as a transparent always-on-top overlay.

- **Song detection** — polls Windows SMTC via PowerShell, picks up title/artist/position from any media source YT Music registers
- **Lyrics sources** — tries LRCLIB first, falls back to NetEase Cloud Music; handles YT Music's messy title formats (channel names, `[Thai Sub]`/`[Viet Sub]`, etc.)
- **Sync** — timestamps from the lyrics file + SMTC position at detection time; recalibrates on seek/pause/resume

## Requirements

- Windows 10/11
- [Go](https://go.dev/) 1.21+
- [Wails CLI](https://wails.io/) v2
- [Node.js](https://nodejs.org/) (for frontend assets)
- MinGW-w64 C compiler (e.g. via [MSYS2](https://www.msys2.org/))

## Getting started

**With clone:**

```bash
git clone https://github.com/mystaline/myrics-overlay
cd myrics-overlay

# Install CLI tool
go install ./cmd/myrics
```

**Without clone:**

```bash
go install github.com/mystaline/myrics-overlay/cmd/myrics@latest
```

---

```bash

# Install dependencies
myrics install-deps

# Copy config
myrics config
# edit configs/config.yaml if you want ACRCloud-based detection (Linux/macOS)

# Build and install overlay
myrics install

# Run
myrics run
```

## Dev mode

```bash
myrics dev   # hot reload, logs to terminal
```

## Config

`configs/config.yaml` (gitignored, copy from `config.yaml.example`):

```yaml
overlay:
  font_size: 18
  position: "top" # top | bottom | center
  opacity: 0.9

lyrics:
  lrclib_url: "https://lrclib.net/api/search"
  netease_search_url: "https://music.163.com/api/search/get"
  netease_lyrics_url: "https://music.163.com/api/song/lyric"

# Linux/macOS only — not needed on Windows
acrcloud:
  access_key: "..."
  secret_key: "..."
  host: "identify-eu-west-1.acrcloud.com"
```

## CLI commands

| Command               | What it does                             |
| --------------------- | ---------------------------------------- |
| `myrics install-deps` | Install Wails CLI + Go deps              |
| `myrics config`       | Copy example config                      |
| `myrics build`        | Build production binary                  |
| `myrics install`      | Build + install overlay to `$GOPATH/bin` |
| `myrics run`          | Launch installed overlay                 |
| `myrics dev`          | Hot-reload dev mode                      |
| `myrics test`         | Run Go tests                             |
| `myrics clean`        | Remove build artifacts                   |
| `myrics uninstall`    | Remove overlay from `$GOPATH/bin`        |

## Gaps / not yet implemented

- **Lyrics sources** — only LRCLIB and NetEase; no Genius or other fallbacks for plain-text lyrics
- **Plain lyrics** — displayed as static first line only; no scrolling or full-text view
- **Non-alphabetical normalization** — Japanese/Korean/Chinese titles from SMTC aren't normalized before querying, which can cause missed matches
- **SMTC session recovery** — if the SMTC session goes stale (e.g. after many skip/pause cycles), requires app restart
- **Windows only** — Linux/macOS path uses ACRCloud for detection but audio capture is not implemented on Windows

## Known limitations

**Sync drift** — SMTC doesn't expose real-time playback position; Myrics estimates it from the position reported at song detection. If lyrics are out of sync, pause and play once — the resume event recalibrates from the exact SMTC position.

## Platform notes

**Windows** — SMTC handles detection. No audio capture needed.

**Linux/macOS** — Uses PortAudio to capture audio and ACRCloud for song recognition. Requires `portaudio-devel` and an ACRCloud API key.
