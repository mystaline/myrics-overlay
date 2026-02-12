package models

import "time"

// SongInfo represents metadata about a detected song
type SongInfo struct {
	Title    string
	Artist   string
	Album    string
	Duration int // in seconds
}

// LyricLine represents a single line of lyrics with timestamp
type LyricLine struct {
	Timestamp time.Duration // offset from song start
	Text      string
}
